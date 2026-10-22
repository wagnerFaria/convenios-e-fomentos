package slug

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	tests := map[string]string{
		"Revisor do Departamento": "revisor-do-departamento",
		"Gestão de Conteúdo":      "gestao-de-conteudo",
		"  Editor (cópia)  ":      "editor-copia",
		"!!!":                     "item",
		"":                        "item",
		"Ação Ñ":                  "acao-n",
	}
	for in, want := range tests {
		assert.Equal(t, want, Make(in), in)
	}
	assert.LessOrEqual(t, len(Make(strings.Repeat("a", 200))), 60)
}
