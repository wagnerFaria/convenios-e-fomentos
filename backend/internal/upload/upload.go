// Pacote upload concentra as regras de arquivos enviados ao armazenamento: tipos aceitos, tamanhos máximos,
// conferência do nome e da assinatura do formato. É usado pelos documentos e pelos anexos dos projetos.
package upload

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// MB é um mebibyte.
const MB = int64(1) << 20

const (
	// MaxFilenameLen é o tamanho máximo do nome do arquivo, em caracteres.
	MaxFilenameLen  = 255
	maxDocumentSize = 200 * MB // PDF e DOCX
	maxImageSize    = 25 * MB
)

// Rule é a regra de um tipo de arquivo aceito.
type Rule struct {
	Exts    []string
	MaxSize int64
	// Magic confere os primeiros bytes: o Content-Type enviado pelo cliente não basta.
	Magic func(head []byte) bool
}

// Rules são os tipos aceitos, por Content-Type.
var Rules = map[string]Rule{
	"application/pdf": {[]string{".pdf"}, maxDocumentSize, func(b []byte) bool { return bytes.HasPrefix(b, []byte("%PDF-")) }},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {[]string{".docx"}, maxDocumentSize,
		func(b []byte) bool { return bytes.HasPrefix(b, []byte("PK\x03\x04")) }},
	"image/png":  {[]string{".png"}, maxImageSize, func(b []byte) bool { return bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")) }},
	"image/jpeg": {[]string{".jpg", ".jpeg"}, maxImageSize, func(b []byte) bool { return bytes.HasPrefix(b, []byte{0xFF, 0xD8, 0xFF}) }},
	"image/webp": {[]string{".webp"}, maxImageSize, func(b []byte) bool {
		return len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP"))
	}},
}

// FieldNames são os nomes dos campos nas mensagens de erro (cada API usa a sua convenção).
type FieldNames struct{ Filename, ContentType, Size string }

// Check valida o arquivo que o cliente pretende enviar e devolve o nome base limpo e as falhas, se houver.
func Check(filename, contentType string, size int64, f FieldNames) (string, []httpapi.FieldError) {
	var issues []httpapi.FieldError
	bad := func(field, m string) { issues = append(issues, httpapi.FieldError{Field: field, Message: m}) }

	// Só o nome base: caminhos enviados pelo cliente não têm significado aqui.
	name := strings.TrimSpace(path.Base(strings.ReplaceAll(filename, `\`, "/")))
	switch {
	case name == "" || name == "." || name == "/":
		bad(f.Filename, "Informe o nome do arquivo.")
	case utf8.RuneCountInString(name) > MaxFilenameLen || !utf8.ValidString(name) || strings.ContainsFunc(name, func(r rune) bool { return r < 0x20 || r == 0x7f }):
		bad(f.Filename, fmt.Sprintf("Nome inválido ou maior que %d caracteres.", MaxFilenameLen))
	}

	rule, ok := Rules[contentType]
	if !ok {
		bad(f.ContentType, "Tipo não permitido. Use PDF, DOCX, PNG, JPEG ou WebP.")
		return name, issues
	}
	if name != "" && !HasExt(name, rule.Exts) {
		bad(f.Filename, "A extensão do arquivo não corresponde ao tipo informado ("+strings.Join(rule.Exts, ", ")+").")
	}
	if size < 1 || size > rule.MaxSize {
		bad(f.Size, fmt.Sprintf("Tamanho deve ficar entre 1 byte e %d MB para este tipo.", rule.MaxSize/MB))
	}
	return name, issues
}

// HasExt informa se o nome tem uma das extensões (sem diferenciar maiúsculas).
func HasExt(name string, exts []string) bool {
	e := strings.ToLower(filepath.Ext(name))
	for _, x := range exts {
		if e == x {
			return true
		}
	}
	return false
}
