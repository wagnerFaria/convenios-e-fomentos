// Pacote mtid isola o MT-ID (assinatura digital do governo de MT, via SmartPass no X-Via) do restante do sistema. O contrato é o
// que a documentação da MTI descreve (docs/documentacao-mti-app-cloud-mti-mt-gov-br-mt-id.md): envio do arquivo, pedido de
// assinatura, consulta, cancelamento e download do arquivo assinado. A assinatura é confirmada pelo signatário no app MTCidadão.
//
// Só existe a versão em memória (Memory): o cliente HTTP real depende de acesso ao X-Via e está em docs/debito-tecnico-mt-id.md.
package mtid

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound indica que o arquivo ou o pedido não existe no MT-ID.
var ErrNotFound = errors.New("mtid: não encontrado")

// State é a situação de um pedido de assinatura.
type State string

const (
	// StatePending: o pedido existe e o signatário ainda não assinou.
	StatePending State = "pending"
	// StateApproved: o signatário assinou; Status.FileID é o arquivo assinado.
	StateApproved State = "approved"
	// StateRejected: o signatário recusou.
	StateRejected State = "rejected"
	// StateExpired: o pedido venceu sem assinatura.
	StateExpired State = "expired"
)

// Request é o pedido de assinatura de um arquivo já enviado.
type Request struct {
	// SignerCPF é o CPF (só dígitos) de quem vai assinar.
	SignerCPF   string
	Description string
	FileID      string
	// Detached: true devolve a assinatura em arquivo separado; false a insere no documento. As declarações usam false: o arquivo
	// assinado é o próprio PDF.
	Detached bool
	// X e Y posicionam a assinatura na página; Pages são as páginas (a primeira é 0, como na documentação do MT-ID).
	X, Y     int
	Pages    []int
	Metadata map[string]string
}

// Status é o resultado da consulta de um pedido.
type Status struct {
	State    State
	SignedAt time.Time
	// FileID é o arquivo assinado, quando State é StateApproved.
	FileID string
}

// Client é o contrato do MT-ID.
type Client interface {
	// Upload envia o arquivo e devolve o seu id (POST /file-upload).
	Upload(ctx context.Context, filename string, pdf []byte) (fileID string, err error)
	// RequestSignature cria o pedido de assinatura e devolve o id do pedido (orderId).
	RequestSignature(ctx context.Context, r Request) (orderID string, err error)
	// Status consulta o pedido (GET /signature/{orderId}).
	Status(ctx context.Context, orderID string) (Status, error)
	// Cancel cancela um pedido pendente (DELETE /signature/{orderId}).
	Cancel(ctx context.Context, orderID string) error
	// Download baixa um arquivo (GET /file-download/{fileId}).
	Download(ctx context.Context, fileID string) ([]byte, error)
}
