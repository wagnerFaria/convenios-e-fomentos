package document

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/storage"
)

// Valores do campo status do arquivo.
const (
	FilePending   = "pendente"
	FileConfirmed = "confirmado"
)

const (
	maxFilenameLen   = 255
	maxFilesPerDoc   = 50
	mb               = int64(1) << 20
	maxDocumentSize  = 200 * mb // PDF e DOCX
	maxImageSize     = 25 * mb
	pendingGraceHour = 24
)

// fileRule é a regra de um tipo de arquivo aceito.
type fileRule struct {
	exts    []string
	maxSize int64
	// magic confere os primeiros bytes: o Content-Type enviado pelo cliente não basta.
	magic func(head []byte) bool
}

var fileRules = map[string]fileRule{
	"application/pdf": {[]string{".pdf"}, maxDocumentSize, func(b []byte) bool { return bytes.HasPrefix(b, []byte("%PDF-")) }},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {[]string{".docx"}, maxDocumentSize,
		func(b []byte) bool { return bytes.HasPrefix(b, []byte("PK\x03\x04")) }},
	"image/png":  {[]string{".png"}, maxImageSize, func(b []byte) bool { return bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")) }},
	"image/jpeg": {[]string{".jpg", ".jpeg"}, maxImageSize, func(b []byte) bool { return bytes.HasPrefix(b, []byte{0xFF, 0xD8, 0xFF}) }},
	"image/webp": {[]string{".webp"}, maxImageSize, func(b []byte) bool {
		return len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP"))
	}},
}

// File é o arquivo anexado a um documento. A chave do objeto nunca sai da API.
type File struct {
	ID          uuid.UUID  `json:"id"`
	DocumentID  uuid.UUID  `json:"document_id"`
	Filename    string     `json:"filename"`
	ContentType string     `json:"content_type"`
	SizeBytes   int64      `json:"size_bytes"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
	objectKey   string
}

func fileFromDB(f db.DocumentFile) File {
	return File{ID: f.ID, DocumentID: f.DocumentID, Filename: f.Filename, ContentType: f.ContentType, SizeBytes: f.SizeBytes,
		Status: f.Status, CreatedAt: f.CreatedAt, ConfirmedAt: f.ConfirmedAt, objectKey: f.ObjectKey}
}

// FileInput descreve o arquivo que o cliente pretende enviar.
type FileInput struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

func (in FileInput) validate() (FileInput, error) {
	var issues []httpapi.FieldError
	bad := func(f, m string) { issues = append(issues, httpapi.FieldError{Field: f, Message: m}) }

	// Só o nome base: caminhos enviados pelo cliente não têm significado aqui.
	name := strings.TrimSpace(path.Base(strings.ReplaceAll(in.Filename, `\`, "/")))
	switch {
	case name == "" || name == "." || name == "/":
		bad("filename", "Informe o nome do arquivo.")
	case utf8.RuneCountInString(name) > maxFilenameLen || !utf8.ValidString(name) || strings.ContainsFunc(name, func(r rune) bool { return r < 0x20 || r == 0x7f }):
		bad("filename", fmt.Sprintf("Nome inválido ou maior que %d caracteres.", maxFilenameLen))
	}
	in.Filename = name

	rule, ok := fileRules[in.ContentType]
	switch {
	case !ok:
		bad("content_type", "Tipo não permitido. Use PDF, DOCX, PNG, JPEG ou WebP.")
	default:
		if name != "" && !hasExt(name, rule.exts) {
			bad("filename", "A extensão do arquivo não corresponde ao tipo informado ("+strings.Join(rule.exts, ", ")+").")
		}
		if in.SizeBytes < 1 || in.SizeBytes > rule.maxSize {
			bad("size_bytes", fmt.Sprintf("Tamanho deve ficar entre 1 byte e %d MB para este tipo.", rule.maxSize/mb))
		}
	}
	if len(issues) > 0 {
		return in, fieldErrors(issues...)
	}
	return in, nil
}

func hasExt(name string, exts []string) bool {
	e := strings.ToLower(filepath.Ext(name))
	for _, x := range exts {
		if e == x {
			return true
		}
	}
	return false
}

var errStorageUnavailable = &httpapi.APIError{Status: 503, Code: "storage_unavailable", Message: "O armazenamento de arquivos não está configurado."}

// SetStorage liga o armazenamento de arquivos. Sem ele, as rotas de arquivos respondem 503.
func (s *Service) SetStorage(store storage.Store, uploadTTL, downloadTTL time.Duration, log *slog.Logger) {
	s.store, s.uploadTTL, s.downloadTTL, s.log = store, uploadTTL, downloadTTL, log
}

func (s *Service) requireStore() error {
	if s.store == nil {
		return errStorageUnavailable
	}
	return nil
}

// authorizedDocument carrega o documento e confere se a decisão da ação cobre ESTE registro (403 se não).
func (s *Service) authorizedDocument(ctx context.Context, id uuid.UUID, dec authz.Decision) (Document, error) {
	d, err := s.Get(ctx, id)
	if err != nil {
		return Document{}, err
	}
	if !dec.AllowsRecord(d.Getter()) {
		return Document{}, httpapi.ErrForbidden()
	}
	return d, nil
}

// publishedDocument carrega o documento só se estiver publicado (404 caso contrário).
func (s *Service) publishedDocument(ctx context.Context, id uuid.UUID) (Document, error) {
	d, err := s.Get(ctx, id)
	if err != nil {
		return Document{}, err
	}
	if d.Status != StatusPublished {
		return Document{}, httpapi.ErrNotFound()
	}
	return d, nil
}

func (s *Service) getFile(ctx context.Context, docID, fileID uuid.UUID) (File, error) {
	f, err := s.q.GetDocumentFile(ctx, db.GetDocumentFileParams{ID: fileID, DocumentID: docID})
	if err != nil {
		return File{}, notFound(err)
	}
	return fileFromDB(f), nil
}

func (s *Service) listFiles(ctx context.Context, docID uuid.UUID, onlyConfirmed bool) ([]File, error) {
	rows, err := s.q.ListDocumentFiles(ctx, db.ListDocumentFilesParams{DocumentID: docID, OnlyConfirmed: onlyConfirmed})
	if err != nil {
		return nil, err
	}
	out := make([]File, len(rows))
	for i, r := range rows {
		out[i] = fileFromDB(r)
	}
	return out, nil
}

// ListFilesByAdmin lista os arquivos (inclusive os pendentes) de um documento coberto por dec.
func (s *Service) ListFilesByAdmin(ctx context.Context, docID uuid.UUID, dec authz.Decision) ([]File, error) {
	if _, err := s.authorizedDocument(ctx, docID, dec); err != nil {
		return nil, err
	}
	return s.listFiles(ctx, docID, false)
}

// ListFilesPublic lista os arquivos confirmados de um documento publicado.
func (s *Service) ListFilesPublic(ctx context.Context, docID uuid.UUID) ([]File, error) {
	if _, err := s.publishedDocument(ctx, docID); err != nil {
		return nil, err
	}
	return s.listFiles(ctx, docID, true)
}

// CreateFile registra o arquivo como pendente e autoriza o envio direto ao armazenamento.
func (s *Service) CreateFile(ctx context.Context, actor AdminActor, docID uuid.UUID, in FileInput, dec authz.Decision) (File, storage.Upload, error) {
	if err := s.requireStore(); err != nil {
		return File{}, storage.Upload{}, err
	}
	in, err := in.validate()
	if err != nil {
		return File{}, storage.Upload{}, err
	}
	if _, err := s.authorizedDocument(ctx, docID, dec); err != nil {
		return File{}, storage.Upload{}, err
	}
	var row db.DocumentFile
	id := uuid.New()
	// A chave é gerada aqui: nada do que o cliente envia entra nela.
	key := fmt.Sprintf("documents/%s/%s", docID, id)
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		// Bloqueia o documento para o limite por documento valer sob concorrência.
		if _, err := q.GetDocumentForUpdate(ctx, docID); err != nil {
			return notFound(err)
		}
		existing, err := q.ListDocumentFiles(ctx, db.ListDocumentFilesParams{DocumentID: docID})
		if err != nil {
			return err
		}
		if len(existing) >= maxFilesPerDoc {
			return httpapi.ErrConflict("too_many_files", fmt.Sprintf("O documento já tem %d arquivos (limite).", maxFilesPerDoc))
		}
		row, err = q.InsertDocumentFile(ctx, db.InsertDocumentFileParams{ID: id, DocumentID: docID, ObjectKey: key, Filename: in.Filename,
			ContentType: in.ContentType, SizeBytes: in.SizeBytes, CreatedByID: &actor.ID})
		return err
	})
	if err != nil {
		return File{}, storage.Upload{}, err
	}
	up, err := s.store.PresignUpload(ctx, key, in.ContentType, fileRules[in.ContentType].maxSize, s.uploadTTL)
	if err != nil {
		_ = s.q.DeleteDocumentFileByID(ctx, id)
		return File{}, storage.Upload{}, err
	}
	return fileFromDB(row), up, nil
}

func unprocessable(field, msg string) error {
	return fieldErrors(httpapi.FieldError{Field: field, Message: msg})
}

// ConfirmFile confere o objeto no armazenamento (existência, tipo, tamanho e assinatura do formato) e
// marca o arquivo como confirmado. Um objeto que não passa é apagado junto com o registro.
func (s *Service) ConfirmFile(ctx context.Context, docID, fileID uuid.UUID, dec authz.Decision) (File, error) {
	if err := s.requireStore(); err != nil {
		return File{}, err
	}
	if _, err := s.authorizedDocument(ctx, docID, dec); err != nil {
		return File{}, err
	}
	f, err := s.getFile(ctx, docID, fileID)
	if err != nil {
		return File{}, err
	}
	if f.Status == FileConfirmed {
		return f, nil
	}
	info, err := s.store.Stat(ctx, f.objectKey)
	if errors.Is(err, storage.ErrNotFound) {
		return File{}, httpapi.ErrConflict("file_not_uploaded", "O arquivo ainda não foi enviado ao armazenamento.")
	}
	if err != nil {
		return File{}, err
	}
	rule := fileRules[f.ContentType]
	reject := func(msg string) (File, error) {
		_ = s.store.Delete(ctx, f.objectKey)
		_ = s.q.DeleteDocumentFileByID(ctx, f.ID)
		return File{}, unprocessable("file", msg)
	}
	if info.Size < 1 || info.Size > rule.maxSize {
		return reject("O arquivo enviado excede o tamanho permitido para o tipo.")
	}
	if !strings.EqualFold(strings.TrimSpace(strings.SplitN(info.ContentType, ";", 2)[0]), f.ContentType) {
		return reject("O tipo do arquivo enviado não corresponde ao informado.")
	}
	head, err := s.store.Head(ctx, f.objectKey, 16)
	if err != nil {
		return File{}, err
	}
	if !rule.magic(head) {
		return reject("O conteúdo do arquivo não corresponde ao formato informado.")
	}
	row, err := s.q.ConfirmDocumentFile(ctx, db.ConfirmDocumentFileParams{ID: f.ID, SizeBytes: info.Size})
	if err != nil {
		return File{}, notFound(err)
	}
	return fileFromDB(row), nil
}

// Download devolve uma URL temporária para baixar o arquivo confirmado.
func (s *Service) download(ctx context.Context, docID, fileID uuid.UUID) (string, time.Time, error) {
	if err := s.requireStore(); err != nil {
		return "", time.Time{}, err
	}
	f, err := s.getFile(ctx, docID, fileID)
	if err != nil {
		return "", time.Time{}, err
	}
	if f.Status != FileConfirmed {
		return "", time.Time{}, httpapi.ErrNotFound()
	}
	return s.store.PresignDownload(ctx, f.objectKey, f.Filename, f.ContentType, s.downloadTTL)
}

// DownloadByAdmin autoriza o download pelo registro do documento.
func (s *Service) DownloadByAdmin(ctx context.Context, docID, fileID uuid.UUID, dec authz.Decision) (string, time.Time, error) {
	if _, err := s.authorizedDocument(ctx, docID, dec); err != nil {
		return "", time.Time{}, err
	}
	return s.download(ctx, docID, fileID)
}

// DownloadPublic só baixa arquivos de documentos publicados.
func (s *Service) DownloadPublic(ctx context.Context, docID, fileID uuid.UUID) (string, time.Time, error) {
	if _, err := s.publishedDocument(ctx, docID); err != nil {
		return "", time.Time{}, err
	}
	return s.download(ctx, docID, fileID)
}

// DeleteFile remove o registro e o objeto.
func (s *Service) DeleteFile(ctx context.Context, docID, fileID uuid.UUID, dec authz.Decision) error {
	if err := s.requireStore(); err != nil {
		return err
	}
	if _, err := s.authorizedDocument(ctx, docID, dec); err != nil {
		return err
	}
	f, err := s.getFile(ctx, docID, fileID)
	if err != nil {
		return err
	}
	// Objeto primeiro: se falhar, o registro continua e a remoção pode ser repetida.
	if err := s.store.Delete(ctx, f.objectKey); err != nil {
		return err
	}
	_, err = s.q.DeleteDocumentFile(ctx, db.DeleteDocumentFileParams{ID: fileID, DocumentID: docID})
	return err
}

// deleteObjects apaga os objetos de um documento já excluído (as linhas saem por ON DELETE CASCADE).
// É melhor esforço: uma falha vira log, pois o documento já não existe.
func (s *Service) deleteObjects(ctx context.Context, keys []string) {
	if s.store == nil {
		return
	}
	for _, k := range keys {
		if err := s.store.Delete(ctx, k); err != nil && s.log != nil {
			s.log.ErrorContext(ctx, "removendo objeto órfão", "key", k, "error", err)
		}
	}
}

// CleanupPendingFiles remove envios que nunca foram confirmados há mais de 24 horas
// (o objeto, se existir, e o registro). Devolve quantos registros removeu.
func (s *Service) CleanupPendingFiles(ctx context.Context) (int, error) {
	if s.store == nil {
		return 0, nil
	}
	rows, err := s.q.ListStalePendingDocumentFiles(ctx, pendingGraceHour)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		if err := s.store.Delete(ctx, r.ObjectKey); err != nil {
			return n, err
		}
		if err := s.q.DeleteDocumentFileByID(ctx, r.ID); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// downloadLink é a resposta dos downloads: a URL temporária assinada.
type downloadLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}
