// Pacote outbox é a caixa de saída de e-mails. Quem muda o estado de algo grava a mensagem na MESMA transação (Enqueue); um
// processo em segundo plano (Worker) a envia depois, com tentativas, sem segurar a requisição e sem perder a mensagem se o SMTP
// estiver fora do ar ou a API cair no meio.
package outbox

import (
	"context"

	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

// Message é um e-mail a enviar.
type Message struct {
	// Kind identifica o tipo do evento (para diagnóstico), por exemplo "project.archived".
	Kind      string
	ProjectID *uuid.UUID
	To        string
	Subject   string
	Body      string
	// DedupKey impede gravar duas vezes o mesmo e-mail para o mesmo destinatário (o evento pode ser reprocessado). Vazio = sempre
	// grava (uma chave aleatória), para quem envia o mesmo tipo de mensagem várias vezes de propósito.
	DedupKey string
}

// Enqueue grava a mensagem usando as consultas q, que devem estar ligadas à transação da mudança de estado: se a transação
// desfizer, o e-mail também não existe. Destinatário vazio é ignorado. Devolve false quando a chave já existia.
func Enqueue(ctx context.Context, q *db.Queries, m Message) (bool, error) {
	if m.To == "" {
		return false, nil
	}
	key := m.DedupKey
	if key == "" {
		key = uuid.NewString()
	}
	n, err := q.EnqueueEmail(ctx, db.EnqueueEmailParams{Kind: m.Kind, ProjectID: m.ProjectID, ToEmail: m.To, Subject: m.Subject, Body: m.Body, DedupKey: key})
	return n > 0, err
}
