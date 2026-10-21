// Pacote slug gera identificadores legíveis (a-z, 0-9 e hífen) a partir de nomes em português.
package slug

import (
	"regexp"
	"strings"
)

var (
	accents = strings.NewReplacer("á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a", "é", "e", "ê", "e", "è", "e", "í", "i", "ì", "i",
		"ó", "o", "õ", "o", "ô", "o", "ò", "o", "ú", "u", "ù", "u", "ü", "u", "ç", "c", "ñ", "n")
	nonSlug = regexp.MustCompile(`[^a-z0-9]+`)
)

const maxLen = 60

// Make converte o nome em slug; nome sem letras ou números vira "item".
func Make(name string) string {
	s := strings.Trim(nonSlug.ReplaceAllString(accents.Replace(strings.ToLower(name)), "-"), "-")
	if s == "" {
		return "item"
	}
	if len(s) > maxLen {
		s = strings.TrimRight(s[:maxLen], "-")
	}
	return s
}
