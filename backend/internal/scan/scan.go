// Pacote scan verifica o conteúdo dos arquivos enviados antes de liberá-los (antivírus).
package scan

import (
	"context"
	"io"
)

// Result é o resultado de uma verificação.
type Result struct {
	// Infected indica que o conteúdo foi marcado como malicioso; Signature é o nome da ameaça.
	Infected  bool
	Signature string
}

// Scanner verifica um fluxo de bytes. Erro significa que NÃO foi possível verificar (serviço fora do ar, estouro de tamanho...):
// quem usa trata como falha e não libera o arquivo.
type Scanner interface {
	Scan(ctx context.Context, r io.Reader) (Result, error)
}
