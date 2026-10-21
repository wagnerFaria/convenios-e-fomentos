package authz

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// Permission é uma permissão de role já carregada.
type Permission struct {
	Action     string
	Subject    string   // vazio = sem recurso
	Fields     []string // nil = todos os campos
	Conditions []string // vazio = sem condição; várias = OR
	Properties map[string]any
}

// Store carrega as permissões de uma role.
type Store interface {
	RolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error)
}

// cache guarda em memória as permissões por role. A geração evita que uma
// leitura iniciada antes de uma invalidação grave dados velhos depois dela.
type cache struct {
	store Store
	mu    sync.RWMutex
	gen   uint64
	roles map[uuid.UUID][]Permission
}

func newCache(store Store) *cache {
	return &cache{store: store, roles: map[uuid.UUID][]Permission{}}
}

func (c *cache) get(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	c.mu.RLock()
	perms, ok := c.roles[roleID]
	gen := c.gen
	c.mu.RUnlock()
	if ok {
		return perms, nil
	}
	perms, err := c.store.RolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.gen == gen {
		c.roles[roleID] = perms
	}
	c.mu.Unlock()
	return perms, nil
}

// invalidate descarta as roles informadas; sem argumentos, descarta tudo.
func (c *cache) invalidate(roleIDs ...uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gen++
	if len(roleIDs) == 0 {
		clear(c.roles)
		return
	}
	for _, id := range roleIDs {
		delete(c.roles, id)
	}
}
