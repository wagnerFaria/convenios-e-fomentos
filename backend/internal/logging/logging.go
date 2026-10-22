// Pacote logging configura o slog em JSON com suporte a request ID.
package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

type ctxKey struct{}

// New cria um logger JSON no nível informado (debug, info, warn, error).
func New(w io.Writer, level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(strings.ToLower(level))); err != nil {
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: l}))
}

// WithRequestID guarda no contexto o ID da requisição.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// RequestID devolve o ID da requisição presente no contexto, se houver.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
