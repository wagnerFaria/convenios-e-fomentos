// Pacote preview converte documentos do Word em PDF para serem exibidos no navegador. A conversão é feita pelo LibreOffice, que roda em
// outro serviço (Gotenberg): o arquivo vem de quem quer que seja, e abri-lo fora da API isola falhas e macros do processo principal.
package preview

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

// ErrUnavailable indica que a conversão não está disponível (serviço fora do ar, recusou o arquivo ou demorou demais).
var ErrUnavailable = errors.New("preview: conversão indisponível")

// Converter transforma um documento do Word em PDF.
type Converter interface {
	// ToPDF converte o conteúdo de src; filename (com a extensão .docx) orienta o formato.
	ToPDF(ctx context.Context, filename string, src io.Reader) ([]byte, error)
}

// maxPDF limita o PDF aceito do serviço, para uma resposta defeituosa não encher a memória.
const maxPDF = 256 << 20

// Gotenberg converte pelo serviço Gotenberg (POST /forms/libreoffice/convert).
type Gotenberg struct {
	baseURL string
	client  *http.Client
}

// NewGotenberg cria o cliente. timeout limita cada conversão.
func NewGotenberg(baseURL string, timeout time.Duration) *Gotenberg {
	return &Gotenberg{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: timeout}}
}

func (g *Gotenberg) ToPDF(ctx context.Context, filename string, src io.Reader) ([]byte, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files"; filename=%q`, "documento.docx"))
		h.Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		part, err := mw.CreatePart(h)
		if err == nil {
			_, err = io.Copy(part, src)
		}
		if err == nil {
			err = mw.Close()
		}
		_ = pw.CloseWithError(err)
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/forms/libreoffice/convert", pr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("%w: status %d: %s", ErrUnavailable, resp.StatusCode, bytes.TrimSpace(msg))
	}
	pdf, err := io.ReadAll(io.LimitReader(resp.Body, maxPDF+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if len(pdf) > maxPDF || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return nil, fmt.Errorf("%w: resposta não é um PDF válido", ErrUnavailable)
	}
	return pdf, nil
}
