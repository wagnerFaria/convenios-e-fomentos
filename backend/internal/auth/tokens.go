package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// NewOpaqueToken gera 32 bytes aleatórios (base64url) e o hash SHA-256 a ser
// persistido. O valor em claro nunca é guardado. SHA-256 basta aqui porque o
// token já tem 256 bits de entropia.
func NewOpaqueToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("gerando token: %w", err)
	}
	plain = base64.RawURLEncoding.EncodeToString(b)
	return plain, HashToken(plain), nil
}

// HashToken calcula o hash persistido de um token opaco.
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
