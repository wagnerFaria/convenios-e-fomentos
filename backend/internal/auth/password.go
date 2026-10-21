// Pacote auth reúne hash de senha, JWT, tokens opacos e a autenticação de administradores.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// maxPasswordLen limita o tamanho da senha para evitar abuso de CPU/memória.
const maxPasswordLen = 128

// Hasher gera e verifica hashes Argon2id no formato PHC.
type Hasher struct {
	Memory      uint32 // KiB
	Iterations  uint32
	Parallelism uint8
	MinLength   int
}

const (
	saltLen = 16
	keyLen  = 32
)

// ErrPasswordPolicy indica que a senha não atende à política mínima.
var ErrPasswordPolicy = errors.New("senha fora da política")

// CheckPolicy valida o tamanho da senha.
func (h Hasher) CheckPolicy(password string) error {
	n := len([]rune(password))
	if n < h.MinLength {
		return fmt.Errorf("%w: mínimo de %d caracteres", ErrPasswordPolicy, h.MinLength)
	}
	if len(password) > maxPasswordLen {
		return fmt.Errorf("%w: máximo de %d caracteres", ErrPasswordPolicy, maxPasswordLen)
	}
	return nil
}

// Hash devolve "$argon2id$v=19$m=..,t=..,p=..$salt$hash".
func (h Hasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("gerando salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, h.Iterations, h.Memory, h.Parallelism, keyLen)
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.Memory, h.Iterations, h.Parallelism, enc.EncodeToString(salt), enc.EncodeToString(key)), nil
}

// Verify compara a senha com o hash em tempo constante. Os parâmetros usados
// são os gravados no próprio hash, então mudar a configuração não quebra senhas antigas.
func (Hasher) Verify(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	enc := base64.RawStdEncoding
	salt, err := enc.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := enc.DecodeString(parts[5])
	if err != nil || len(want) == 0 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(want))) //nolint:gosec // len(want) é pequeno
	return subtle.ConstantTimeCompare(got, want) == 1
}
