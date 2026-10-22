package enduser

import (
	"context"
	"fmt"
	"sync"
)

// ProviderLocal é o provedor padrão (e-mail/usuário + senha), tratado diretamente pelo Service.
const ProviderLocal = "local"

// Identity é a identidade confirmada por um provedor externo (OAuth/OIDC, futuramente).
type Identity struct {
	Provider      string // nome do provedor
	Subject       string // identificador do usuário no provedor
	Email         string
	Username      string
	EmailVerified bool
}

// Provider é a extensão para provedores de identidade externos. Nenhum está implementado
// hoje: a estrutura existe para que adicionar OAuth não exija mudar o modelo de dados
// (end_users.provider) nem as rotas. Um provedor recebe as credenciais da etapa de callback
// (ex.: code e state) e devolve a identidade confirmada.
type Provider interface {
	Name() string
	Authenticate(ctx context.Context, credentials map[string]string) (Identity, error)
}

// Providers é o registro de provedores externos.
type Providers struct {
	mu sync.RWMutex
	m  map[string]Provider
}

// NewProviders cria um registro vazio.
func NewProviders() *Providers { return &Providers{m: map[string]Provider{}} }

// Register adiciona um provedor. O nome "local" é reservado.
func (p *Providers) Register(pr Provider) error {
	name := pr.Name()
	if name == "" || name == ProviderLocal {
		return fmt.Errorf("nome de provedor inválido ou reservado: %q", name)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, dup := p.m[name]; dup {
		return fmt.Errorf("provedor %q já registrado", name)
	}
	p.m[name] = pr
	return nil
}

// Get busca um provedor pelo nome.
func (p *Providers) Get(name string) (Provider, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	pr, ok := p.m[name]
	return pr, ok
}
