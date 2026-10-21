package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testHasher = Hasher{Memory: 64, Iterations: 1, Parallelism: 1, MinLength: 8}

func TestHasher(t *testing.T) {
	h, err := testHasher.Hash("minha-senha-forte")
	require.NoError(t, err)
	assert.Contains(t, h, "$argon2id$v=19$m=64,t=1,p=1$")
	assert.True(t, testHasher.Verify("minha-senha-forte", h))
	assert.False(t, testHasher.Verify("outra-senha", h))
	assert.False(t, testHasher.Verify("x", "lixo"))
	assert.False(t, testHasher.Verify("x", ""))

	// Salt aleatório: o mesmo texto gera hashes diferentes.
	h2, _ := testHasher.Hash("minha-senha-forte")
	assert.NotEqual(t, h, h2)

	// Hashes antigos continuam válidos se os parâmetros da config mudarem.
	other := Hasher{Memory: 128, Iterations: 2, Parallelism: 1}
	assert.True(t, other.Verify("minha-senha-forte", h))
}

func TestPasswordPolicy(t *testing.T) {
	tests := []struct {
		nome  string
		senha string
		erro  bool
	}{
		{"curta", "1234567", true},
		{"mínima", "12345678", false},
		{"unicode conta caracteres", "àéíõüçãô", false},
		{"longa demais", string(make([]byte, 129)), true},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			err := testHasher.CheckPolicy(tt.senha)
			if tt.erro {
				assert.ErrorIs(t, err, ErrPasswordPolicy)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOpaqueToken(t *testing.T) {
	p1, h1, err := NewOpaqueToken()
	require.NoError(t, err)
	p2, _, _ := NewOpaqueToken()
	assert.NotEqual(t, p1, p2)
	assert.Equal(t, h1, HashToken(p1))
	assert.NotContains(t, h1, p1)
	assert.Len(t, h1, 64)
}

func TestJWT(t *testing.T) {
	const secret = "segredo-de-teste-com-mais-de-32-caracteres"
	admin := NewJWTIssuer(secret, AudienceAPI, time.Minute)
	sub, sid := uuid.New(), uuid.New()

	tok, err := admin.Issue(sub, sid)
	require.NoError(t, err)
	c, err := admin.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, sid, c.SessionID)
	got, _ := c.SubjectID()
	assert.Equal(t, sub, got)

	t.Run("audiência errada", func(t *testing.T) {
		api := NewJWTIssuer(secret, "outra-audiencia", time.Minute)
		_, err := api.Parse(tok)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
	t.Run("segredo errado", func(t *testing.T) {
		_, err := NewJWTIssuer("outro-segredo-de-teste-com-32-caracteres", AudienceAPI, time.Minute).Parse(tok)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
	t.Run("expirado", func(t *testing.T) {
		exp := NewJWTIssuer(secret, AudienceAPI, time.Minute)
		exp.now = func() time.Time { return time.Now().Add(-time.Hour) }
		old, _ := exp.Issue(sub, sid)
		_, err := admin.Parse(old)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
	t.Run("algoritmo none", func(t *testing.T) {
		none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer, Subject: sub.String(), Audience: jwt.ClaimStrings{AudienceAPI}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}}).SignedString(jwt.UnsafeAllowNoneSignatureType)
		_, err := admin.Parse(none)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
	t.Run("lixo", func(t *testing.T) {
		_, err := admin.Parse("abc")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
