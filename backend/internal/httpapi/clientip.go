package httpapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type clientIPKey struct{}

// ParseCIDRs converte uma lista de CIDRs (ou IPs avulsos) de proxies confiáveis.
func ParseCIDRs(list []string) ([]*net.IPNet, error) {
	var nets []*net.IPNet
	for _, s := range list {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.Contains(s, "/") {
			if strings.Contains(s, ":") {
				s += "/128"
			} else {
				s += "/32"
			}
		}
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("TRUSTED_PROXIES inválido %q: %w", s, err)
		}
		nets = append(nets, n)
	}
	return nets, nil
}

func inNets(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// resolveClientIP usa o endereço da conexão. Só quando ele pertence a um proxy
// confiável o X-Forwarded-For é considerado: percorre-o da direita para a
// esquerda e devolve o primeiro endereço que não seja de proxy confiável.
func resolveClientIP(r *http.Request, trusted []*net.IPNet) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil || !inNets(ip, trusted) {
		return host
	}
	parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	for i := len(parts) - 1; i >= 0; i-- {
		p := net.ParseIP(strings.TrimSpace(parts[i]))
		if p == nil {
			continue
		}
		if !inNets(p, trusted) {
			return p.String()
		}
	}
	return host
}

// ClientIPMiddleware guarda o IP real do cliente no contexto.
func ClientIPMiddleware(trusted []*net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), clientIPKey{}, resolveClientIP(r, trusted))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClientIP devolve o IP do cliente calculado pelo ClientIPMiddleware.
func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}
