package app_test

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/wagner/convenios-e-fomentos/backend/api"
)

// A especificação OpenAPI precisa descrever exatamente as rotas implementadas.
func TestOpenAPIEstaCoerenteComAsRotas(t *testing.T) {
	e := newEnv(t)

	implemented := map[string]bool{}
	require.NoError(t, chi.Walk(e.app.Router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		route = strings.TrimSuffix(route, "/")
		if route == "" {
			route = "/"
		}
		// A UI e o arquivo da especificação não fazem parte da especificação.
		if route == "/docs" || route == "/docs/openapi.yaml" {
			return nil
		}
		implemented[strings.ToLower(method)+" "+route] = true
		return nil
	}))

	var spec struct {
		OpenAPI string `yaml:"openapi"`
		Paths   map[string]map[string]struct {
			Summary   string                    `yaml:"summary"`
			Responses map[string]map[string]any `yaml:"responses"`
			Tags      []string                  `yaml:"tags"`
		} `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(api.OpenAPI, &spec))
	assert.True(t, strings.HasPrefix(spec.OpenAPI, "3.1"))

	documented := map[string]bool{}
	for path, ops := range spec.Paths {
		for method, op := range ops {
			documented[method+" "+path] = true
			assert.NotEmpty(t, op.Summary, "%s %s sem summary", method, path)
			assert.NotEmpty(t, op.Tags, "%s %s sem tags", method, path)
			assert.NotEmpty(t, op.Responses, "%s %s sem respostas", method, path)
		}
	}

	var missing, extra []string
	for k := range implemented {
		if !documented[k] {
			missing = append(missing, k)
		}
	}
	for k := range documented {
		if !implemented[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	assert.Empty(t, missing, "rotas implementadas e ausentes do openapi.yaml")
	assert.Empty(t, extra, "rotas do openapi.yaml que não existem")
	t.Logf("%d operações documentadas", len(documented))
}
