package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitByIPAndField limita requisições por minuto por IP + o valor de um
// campo JSON do corpo (ex.: "email"). Se o campo não existir, limita só por IP.
// O corpo é lido (já limitado pelo BodyLimit) e restaurado para o handler.
func RateLimitByIPAndField(perMinute int, field string) func(http.Handler) http.Handler {
	key := func(r *http.Request) (string, error) {
		ip := ClientIP(r.Context())
		if r.Body == nil {
			return ip, nil
		}
		raw, err := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(raw))
		if err != nil {
			return ip, nil //nolint:nilerr // falha de leitura é tratada pelo handler
		}
		var m map[string]any
		if json.Unmarshal(raw, &m) == nil {
			if v, ok := m[field].(string); ok && v != "" {
				return ip + "|" + strings.ToLower(v), nil
			}
		}
		return ip, nil
	}
	return httprate.LimitBy(perMinute, time.Minute, key,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			Error(w, r, ErrTooManyRequests())
		}),
	)
}

// RateLimitByIP limita as requisições por minuto de cada IP (já resolvido pelo
// ClientIPMiddleware). perMinute <= 0 desliga o limite.
func RateLimitByIP(perMinute int) func(http.Handler) http.Handler {
	if perMinute <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return httprate.LimitBy(perMinute, time.Minute,
		func(r *http.Request) (string, error) { return ClientIP(r.Context()), nil },
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) { Error(w, r, ErrTooManyRequests()) }),
	)
}
