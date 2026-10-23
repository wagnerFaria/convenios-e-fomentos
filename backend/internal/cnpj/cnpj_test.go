package cnpj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"11.222.333/0001-81":  "11222333000181",
		"11222333000181":      "11222333000181",
		" 11 222 333 0001 81": "11222333000181",
		"00.000.000/0001-91":  "00000000000191", // CNPJ do Banco do Brasil
	} {
		got, err := Normalize(in)
		assert.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
}

func TestNormalizeRecusaInvalidos(t *testing.T) {
	for _, in := range []string{
		"", "123", "11222333000182", // dígito errado
		"11111111111111",     // repetidos
		"11.222.333/0001-8A", // letra
		"112223330001810",    // longo demais
		"00000000000000",     // zeros
	} {
		_, err := Normalize(in)
		assert.ErrorIs(t, err, ErrInvalid, in)
	}
}

func TestFormat(t *testing.T) {
	assert.Equal(t, "11.222.333/0001-81", Format("11222333000181"))
	assert.Equal(t, "123", Format("123"))
}
