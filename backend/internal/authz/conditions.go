package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Nomes lógicos de campos usados pelas condições padrão. Os recursos que
// queiram usá-las devem expor campos com estes nomes no Schema.
const (
	FieldCreatedByID  = "created_by_id"
	FieldDepartmentID = "department_id"
)

// Result é o resultado de uma condição: permite tudo, nega, ou restringe por filtro.
type Result struct {
	filter Filter
}

// Allow permite sem filtro.
func Allow() Result { return Result{filter: FilterTrue()} }

// Deny nega.
func Deny() Result { return Result{filter: FilterFalse()} }

// Restrict permite apenas os registros que casam com o filtro.
func Restrict(f Filter) Result { return Result{filter: f} }

// Handler avalia uma condição para o sujeito autenticado.
type Handler func(ctx context.Context, s Subject) (Result, error)

// Condition é uma condição registrada em código (Condition Registry).
type Condition struct {
	ID       string
	Name     string // nome legível, em português
	Category string
	Handler  Handler
}

// ConditionRegistry guarda as condições disponíveis.
type ConditionRegistry struct {
	conds map[string]Condition
	order []string
}

// NewConditionRegistry cria um registro vazio.
func NewConditionRegistry() *ConditionRegistry {
	return &ConditionRegistry{conds: map[string]Condition{}}
}

// Register adiciona uma condição.
func (r *ConditionRegistry) Register(c Condition) error {
	if c.ID == "" || c.Name == "" || c.Handler == nil {
		return fmt.Errorf("condição %q: id, nome e handler são obrigatórios", c.ID)
	}
	if _, dup := r.conds[c.ID]; dup {
		return fmt.Errorf("condição %q já registrada", c.ID)
	}
	r.conds[c.ID] = c
	r.order = append(r.order, c.ID)
	return nil
}

// MustRegister registra ou entra em pânico; para uso na inicialização.
func (r *ConditionRegistry) MustRegister(cs ...Condition) {
	for _, c := range cs {
		if err := r.Register(c); err != nil {
			panic(err)
		}
	}
}

// Get busca uma condição.
func (r *ConditionRegistry) Get(id string) (Condition, bool) { c, ok := r.conds[id]; return c, ok }

// All lista as condições na ordem de registro.
func (r *ConditionRegistry) All() []Condition {
	out := make([]Condition, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.conds[id])
	}
	return out
}

// RoleMembers consulta administradores que pertencem a alguma das roles.
type RoleMembers interface {
	AdminIDsWithAnyRole(ctx context.Context, roleIDs []uuid.UUID) ([]uuid.UUID, error)
}

// RegisterStandardConditions registra as condições padrão (admin::is-creator e
// admin::has-same-role-as-creator) e o exemplo customizado admin::same-department.
func RegisterStandardConditions(r *ConditionRegistry, members RoleMembers) {
	r.MustRegister(
		Condition{
			ID: "admin::is-creator", Name: "É o criador", Category: "default",
			Handler: func(_ context.Context, s Subject) (Result, error) {
				return Restrict(Eq(FieldCreatedByID, s.ID)), nil
			},
		},
		Condition{
			ID: "admin::has-same-role-as-creator", Name: "Tem a mesma role do criador", Category: "default",
			Handler: func(ctx context.Context, s Subject) (Result, error) {
				ids, err := members.AdminIDsWithAnyRole(ctx, s.RoleIDs())
				if err != nil {
					return Deny(), err
				}
				vals := make([]any, len(ids))
				for i, id := range ids {
					vals[i] = id
				}
				// A lista é resolvida antes da consulta, o que mantém o filtro simples e parametrizado.
				return Restrict(In(FieldCreatedByID, vals...)), nil
			},
		},
		Condition{
			ID: "admin::same-department", Name: "Pertence ao mesmo departamento", Category: "custom",
			Handler: func(_ context.Context, s Subject) (Result, error) {
				if s.DepartmentID == nil {
					return Deny(), nil
				}
				return Restrict(Eq(FieldDepartmentID, *s.DepartmentID)), nil
			},
		},
	)
}
