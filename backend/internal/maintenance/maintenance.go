// Pacote maintenance reúne rotinas de limpeza de dados que só crescem (sessões, tokens de uso único e e-mails já resolvidos).
package maintenance

import (
	"context"
	"fmt"
	"time"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

// Result informa quantas linhas cada rotina removeu.
type Result struct {
	EndUserSessions int64
	OneTimeTokens   int64
	// ResolvedEmails são os e-mails da caixa de saída já enviados (ou desistidos). O corpo tem dados pessoais, então não se guarda.
	ResolvedEmails int64
}

// Cleanup remove sessões e tokens de uso único vencidos, e e-mails resolvidos, há mais de graceDays dias.
// É seguro executar a qualquer momento e quantas vezes for preciso (idempotente).
func Cleanup(ctx context.Context, q *db.Queries, graceDays int) (Result, error) {
	if graceDays < 0 {
		return Result{}, fmt.Errorf("graceDays deve ser >= 0")
	}
	days := int32(graceDays) //nolint:gosec // validado acima; valores absurdos só deixam a limpeza mais conservadora
	var r Result
	var err error
	if r.EndUserSessions, err = q.DeleteExpiredEndUserSessions(ctx, days); err != nil {
		return r, fmt.Errorf("sessões de usuário final: %w", err)
	}
	if r.OneTimeTokens, err = q.DeleteStaleOneTimeTokens(ctx, days); err != nil {
		return r, fmt.Errorf("tokens de uso único: %w", err)
	}
	if r.ResolvedEmails, err = q.PurgeResolvedEmails(ctx, time.Now().AddDate(0, 0, -graceDays)); err != nil {
		return r, fmt.Errorf("e-mails já enviados: %w", err)
	}
	return r, nil
}
