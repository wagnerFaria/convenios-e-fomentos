// Pacote cnpj normaliza, valida e formata números de CNPJ.
package cnpj

import (
	"errors"
	"strings"
)

// ErrInvalid indica um CNPJ malformado ou com dígitos verificadores incorretos.
var ErrInvalid = errors.New("CNPJ inválido")

const length = 14

// Normalize remove a pontuação e valida o CNPJ, devolvendo só os 14 dígitos.
func Normalize(s string) (string, error) {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '-' || r == '/' || r == ' ':
		default:
			return "", ErrInvalid
		}
	}
	digits := b.String()
	if !valid(digits) {
		return "", ErrInvalid
	}
	return digits, nil
}

func valid(d string) bool {
	if len(d) != length || strings.Count(d, string(d[0])) == length {
		return false
	}
	return checkDigit(d[:12]) == int(d[12]-'0') && checkDigit(d[:13]) == int(d[13]-'0')
}

// checkDigit calcula o dígito verificador dos primeiros len(d) dígitos (pesos 2..9, da direita para a esquerda).
func checkDigit(d string) int {
	sum, weight := 0, 2
	for i := len(d) - 1; i >= 0; i-- {
		sum += int(d[i]-'0') * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}

// Format devolve o CNPJ no formato 12.345.678/0001-95. Entradas que não têm 14 dígitos voltam inalteradas.
func Format(digits string) string {
	if len(digits) != length {
		return digits
	}
	return digits[:2] + "." + digits[2:5] + "." + digits[5:8] + "/" + digits[8:12] + "-" + digits[12:]
}
