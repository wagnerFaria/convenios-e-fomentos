package cpf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		nome, in, want string
		erro           bool
	}{
		{"só dígitos", "52998224725", "52998224725", false},
		{"com pontuação", "529.982.247-25", "52998224725", false},
		{"com espaços", " 529.982.247-25 ", "52998224725", false},
		{"dígito verificador errado", "52998224726", "", true},
		{"sequência repetida", "111.111.111-11", "", true},
		{"curto", "5299822472", "", true},
		{"longo", "529982247250", "", true},
		{"letras", "5299822472a", "", true},
		{"vazio", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			got, err := Normalize(tt.in)
			if tt.erro {
				require.ErrorIs(t, err, ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatEMask(t *testing.T) {
	assert.Equal(t, "529.982.247-25", Format("52998224725"))
	assert.Equal(t, "***.982.247-**", Mask("52998224725"))
	assert.Equal(t, "123", Format("123"))
	assert.Empty(t, Mask("123"))
}
