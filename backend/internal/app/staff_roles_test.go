package app_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/seed"
)

// Os perfis de equipe nascem nas migrações; o seed declara os mesmos valores. Este teste impede que os dois divirjam.
func TestPerfisDeEquipeDaMigracaoBatemComOSeed(t *testing.T) {
	e, _, _ := seeded(t)

	for _, code := range []string{seed.RoleAnalyst, seed.RoleSupervisor, seed.RoleReviewer, seed.RoleSecretary} {
		t.Run(code, func(t *testing.T) {
			rows, err := e.pool.Query(context.Background(),
				`SELECT p.action, p.conditions::text, coalesce(p.fields::text, '') FROM admin_permissions p
				 JOIN admin_roles r ON r.id = p.role_id WHERE r.code = $1`, code)
			require.NoError(t, err)
			defer rows.Close()
			got := map[string][]string{}
			for rows.Next() {
				var action, conds, fields string
				require.NoError(t, rows.Scan(&action, &conds, &fields))
				assert.Empty(t, fields, "%s: perfis de equipe não restringem campos", action)
				var list []string
				require.NoError(t, json.Unmarshal([]byte(conds), &list))
				got[action] = list
			}
			require.NoError(t, rows.Err())

			want := map[string][]string{}
			for _, p := range seed.DefaultPermissions(code) {
				conds := p.Conditions
				if conds == nil {
					conds = []string{}
				}
				want[p.Action] = conds
			}
			require.NotEmpty(t, want, "o seed declara o perfil %s", code)
			assert.Equal(t, want, got, "migração e seed devem concordar")
		})
	}
}

// As ações de decisão ficam com quem decide: o Supervisor só acompanha, o Analista devolve e arquiva, o Secretário nega.
func TestQuemPodeDecidirEmProjetos(t *testing.T) {
	e, _, _ := seeded(t)
	has := func(code, action string) bool {
		var n int
		require.NoError(t, e.pool.QueryRow(context.Background(),
			`SELECT count(*) FROM admin_permissions p JOIN admin_roles r ON r.id = p.role_id WHERE r.code = $1 AND p.action = $2`, code, action).Scan(&n))
		return n > 0
	}
	table := []struct {
		role   string
		action string
		want   bool
	}{
		{"administrative-supervisor", "admin::projects.read", true},
		{"administrative-supervisor", "admin::projects.request-complement", false},
		{"administrative-supervisor", "admin::projects.archive", false},
		{"administrative-supervisor", "admin::projects.deny", false},
		{"analyst", "admin::projects.request-complement", true},
		{"analyst", "admin::projects.archive", true},
		{"analyst", "admin::projects.deny", false},
		{"analyst", "admin::projects.start-analysis", true},
		{"analyst", "admin::projects.review-item", true},
		{"administrative-supervisor", "admin::projects.start-analysis", false},
		{"administrative-supervisor", "admin::projects.review-item", false},
		{"state-secretary", "admin::projects.start-analysis", false},
		{"state-secretary", "admin::projects.review-item", false},
		{"department-reviewer", "admin::projects.start-analysis", false},
		{"department-reviewer", "admin::projects.review-item", false},
		{"analyst", "admin::settings.departments.read", true},
		{"analyst", "admin::settings.departments.update", true},
		{"analyst", "admin::settings.departments.delete", true},
		{"analyst", "admin::settings.departments.create", false},
		{"state-secretary", "admin::projects.read", true},
		{"state-secretary", "admin::projects.deny", true},
		{"state-secretary", "admin::projects.archive", false},
		{"state-secretary", "admin::projects.request-complement", false},
		{"department-reviewer", "admin::projects.read", true},
		{"department-reviewer", "admin::projects.deny", false},
		{"department-reviewer", "admin::content.document.read", false},
	}
	for _, c := range table {
		assert.Equal(t, c.want, has(c.role, c.action), "%s / %s", c.role, c.action)
	}
}
