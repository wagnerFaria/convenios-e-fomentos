package project

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/preview"
)

// Exibição dos documentos no navegador: a equipe confere o arquivo sem baixá-lo. PDF e imagem são exibidos como estão; o Word é
// convertido para PDF (uma vez, guardado ao lado do original) pelo serviço de conversão. O download continua entregando o original.

const (
	docxType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	// maxConvertSize é o maior Word convertido: acima disso a conversão só consome memória e tempo, e o arquivo é baixado.
	maxConvertSize = 50 << 20

	PreviewPDF   = "pdf"
	PreviewImage = "image"
)

// SetPreview liga a conversão de Word para PDF e a validade das URLs de exibição. Sem conversor, só PDF e imagem são exibidos.
func (s *Service) SetPreview(c preview.Converter, ttl time.Duration, log *slog.Logger) {
	s.converter, s.previewTTL = c, ttl
	if s.log == nil {
		s.log = log
	}
}

// PreviewLink é a URL temporária para exibir um arquivo no navegador.
type PreviewLink struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
	// Kind diz como exibir: pdf ou image.
	Kind        string `json:"kind"`
	ContentType string `json:"contentType"`
	// Converted: o que se vê é o PDF gerado a partir do Word, não o arquivo enviado.
	Converted bool `json:"converted"`
}

func errPreviewUnavailable(msg string) error {
	return &httpapi.APIError{Status: 422, Code: "preview_unavailable", Message: msg}
}

// previewKey é onde fica o PDF gerado de um arquivo do Word.
func previewKey(fileID uuid.UUID) string { return "previews/" + fileID.String() + ".pdf" }

var convertGroup singleflight.Group

// Preview devolve a URL para exibir o arquivo confirmado (404 se a leitura não cobre o projeto). Todo acesso fica na auditoria.
func (s *Service) Preview(ctx context.Context, actor Actor, id, fileID uuid.UUID, read authz.Decision) (PreviewLink, error) {
	if err := s.requireStore(); err != nil {
		return PreviewLink{}, err
	}
	f, err := s.projectFile(ctx, id, fileID, read, httpapi.ErrNotFound)
	if err != nil {
		return PreviewLink{}, err
	}
	if f.Status != FileConfirmed {
		return PreviewLink{}, httpapi.ErrNotFound()
	}
	ttl := s.previewTTL
	if ttl <= 0 {
		ttl = s.downloadTTL
	}
	link := PreviewLink{}
	key := f.ObjectKey
	switch {
	case f.ContentType == "application/pdf":
		link.Kind, link.ContentType = PreviewPDF, f.ContentType
	case len(f.ContentType) > 6 && f.ContentType[:6] == "image/":
		link.Kind, link.ContentType = PreviewImage, f.ContentType
	case f.ContentType == docxType:
		if key, err = s.convertedKey(ctx, f.ObjectKey, f.ID, f.SizeBytes); err != nil {
			return PreviewLink{}, err
		}
		link.Kind, link.ContentType, link.Converted = PreviewPDF, "application/pdf", true
	default:
		return PreviewLink{}, errPreviewUnavailable("Este tipo de arquivo não pode ser exibido. Baixe-o para conferir.")
	}
	link.URL, link.ExpiresAt, err = s.store.PresignView(ctx, key, link.ContentType, ttl)
	if err != nil {
		return PreviewLink{}, err
	}
	if err := s.audit.Record(ctx, s.q, actor.entry("project.file-preview", id, map[string]any{"file": fileID, "document": str(f.RequirementCode)})); err != nil {
		return PreviewLink{}, err
	}
	return link, nil
}

// convertedKey garante que exista o PDF do arquivo do Word e devolve sua chave. Conversões simultâneas do mesmo arquivo viram uma só.
func (s *Service) convertedKey(ctx context.Context, objectKey string, fileID uuid.UUID, size int64) (string, error) {
	key := previewKey(fileID)
	if _, err := s.store.Stat(ctx, key); err == nil {
		return key, nil
	}
	if s.converter == nil {
		return "", errPreviewUnavailable("A conversão de documentos do Word não está configurada. Baixe o arquivo para conferir.")
	}
	if size > maxConvertSize {
		return "", errPreviewUnavailable("O documento é grande demais para ser exibido. Baixe o arquivo para conferir.")
	}
	_, err, _ := convertGroup.Do(key, func() (any, error) {
		// A conversão não pode morrer com o pedido que a iniciou: outros esperam o mesmo resultado.
		cctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Minute)
		defer cancel()
		src, err := s.store.Get(cctx, objectKey)
		if err != nil {
			return nil, err
		}
		defer func() { _ = src.Close() }()
		pdf, err := s.converter.ToPDF(cctx, "documento.docx", io.LimitReader(src, maxConvertSize+1))
		if err != nil {
			return nil, err
		}
		return nil, s.store.PutObject(cctx, key, "application/pdf", pdf)
	})
	if err != nil {
		if errors.Is(err, preview.ErrUnavailable) {
			if s.log != nil {
				s.log.Warn("conversão do documento falhou", "file", fileID, "error", err)
			}
			return "", errPreviewUnavailable("Não foi possível converter o documento para exibição. Baixe o arquivo para conferir.")
		}
		return "", err
	}
	return key, nil
}

// MarkViewed registra que a pessoa da equipe viu o arquivo no visualizador. É a prova exigida para aprovar ou reprovar o documento; o
// titular e os membros do projeto não geram registro. O cliente chama depois de exibir o conteúdo, não ao pedir a URL.
func (s *Service) MarkViewed(ctx context.Context, actor Actor, id, fileID uuid.UUID, read authz.Decision) error {
	f, err := s.projectFile(ctx, id, fileID, read, httpapi.ErrNotFound)
	if err != nil {
		return err
	}
	if f.Status != FileConfirmed {
		return httpapi.ErrNotFound()
	}
	return s.recordStaffOpening(ctx, id, fileID, actor)
}
