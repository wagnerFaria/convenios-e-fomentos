package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/wagner/convenios-e-fomentos/backend/internal/logging"
)

const requestIDHeader = "X-Request-ID"

// Só aceita IDs simples vindos do cliente; qualquer outra coisa é substituída (evita lixo nos logs e nos cabeçalhos).
var requestIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID gera (ou reaproveita, se for válido) o ID da requisição.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if !requestIDRe.MatchString(id) {
			b := make([]byte, 12)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(logging.WithRequestID(r.Context(), id)))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// AccessLog registra uma linha por requisição. Nunca registra corpo, query
// string nem cabeçalhos, para não vazar senhas ou tokens.
func AccessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.InfoContext(r.Context(), "requisição",
				"request_id", logging.RequestID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// Recoverer converte pânicos em erro 500 padronizado.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					log.ErrorContext(r.Context(), "pânico na requisição",
						"panic", p, "stack", string(debug.Stack()),
						"request_id", logging.RequestID(r.Context()))
					Error(w, r, &APIError{Status: 500, Code: "internal_error", Message: "Erro interno do servidor.",
						Details: map[string]string{"requestId": logging.RequestID(r.Context())}})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adiciona cabeçalhos de segurança básicos.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		// As respostas são dinâmicas e podem conter dados pessoais: nada de cache compartilhado.
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// RejectNUL recusa requisições com o caractere NUL na URL ou na query: o PostgreSQL não guarda NUL em texto
// e ele nunca é válido em um parâmetro legítimo.
func RejectNUL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "\x00") || strings.Contains(r.URL.RawQuery, "%00") || strings.Contains(r.URL.RawQuery, "\x00") {
			Error(w, r, ErrBadRequest("A requisição contém caracteres inválidos."))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// BodyLimit limita o tamanho do corpo das requisições.
func BodyLimit(max int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, max)
			}
			next.ServeHTTP(w, r)
		})
	}
}
