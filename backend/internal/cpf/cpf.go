// Pacote cpf normaliza, valida e formata números de CPF.
package cpf

import (
	"errors"
	"strings"
)

// ErrInvalid indica um CPF malformado ou com dígitos verificadores incorretos.
var ErrInvalid = errors.New("CPF inválido")

const length = 11

// Normalize remove a pontuação e valida o CPF, devolvendo só os 11 dígitos.
func Normalize(s string) (string, error) {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '-' || r == ' ':
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
	return checkDigit(d[:9]) == int(d[9]-'0') && checkDigit(d[:10]) == int(d[10]-'0')
}

// checkDigit calcula o dígito verificador dos primeiros len(d) dígitos.
func checkDigit(d string) int {
	sum := 0
	for i := range len(d) {
		sum += int(d[i]-'0') * (len(d) + 1 - i)
	}
	r := (sum * 10) % 11
	if r == 10 {
		return 0
	}
	return r
}

// Format devolve o CPF no formato 123.456.789-09. Entradas que não têm 11 dígitos voltam inalteradas.
func Format(digits string) string {
	if len(digits) != length {
		return digits
	}
	return digits[:3] + "." + digits[3:6] + "." + digits[6:9] + "-" + digits[9:]
}

// Mask devolve o CPF com os dígitos iniciais e finais ocultos (***.456.789-**),
// para listagens, logs e auditoria.
func Mask(digits string) string {
	if len(digits) != length {
		return ""
	}
	return "***." + digits[3:6] + "." + digits[6:9] + "-**"
}
