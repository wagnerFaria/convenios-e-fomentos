package auth_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/testutil"
)

const cpfValido = "529.982.247-25"

func newVerifier(t *testing.T, f *testutil.FakeOIDC) *auth.OIDCVerifier {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	v, err := auth.NewOIDCVerifier(ctx, auth.OIDCConfig{JWKSURI: f.JWKSURI(), Issuer: f.Issuer, ClientID: f.ClientID, CPFClaim: f.CPFClaim},
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	require.NoError(t, err)
	return v
}

func TestOIDCVerifyAceitaTokenValido(t *testing.T) {
	f := testutil.NewFakeOIDC(t)
	v := newVerifier(t, f)

	id, err := v.Verify(f.Token(testutil.Identity{CPF: cpfValido, Email: "ana@exemplo.com", Given: "Ana", Family: "da Silva"}))
	require.NoError(t, err)
	assert.Equal(t, "52998224725", id.CPF, "CPF normalizado para dígitos")
	assert.Equal(t, "ana@exemplo.com", id.Email)
	assert.Equal(t, "Ana", id.FirstName)
	assert.Equal(t, "da Silva", id.LastName)
}

func TestOIDCVerifyNomeAPartirDoClaimName(t *testing.T) {
	f := testutil.NewFakeOIDC(t)
	v := newVerifier(t, f)

	tok := f.TokenWith(testutil.Identity{CPF: cpfValido}, func(c jwt.MapClaims) {
		delete(c, "given_name")
		delete(c, "family_name")
		c["name"] = "Maria Souza Lima"
	})
	id, err := v.Verify(tok)
	require.NoError(t, err)
	assert.Equal(t, "Maria", id.FirstName)
	assert.Equal(t, "Souza Lima", id.LastName)

	tok = f.TokenWith(testutil.Identity{CPF: cpfValido}, func(c jwt.MapClaims) { delete(c, "given_name"); delete(c, "family_name") })
	id, err = v.Verify(tok)
	require.NoError(t, err)
	assert.Equal(t, "Usuário", id.FirstName, "sem nome algum, usa um nome genérico")
}

func TestOIDCVerifyRejeita(t *testing.T) {
	f := testutil.NewFakeOIDC(t)
	v := newVerifier(t, f)
	user := testutil.Identity{CPF: cpfValido, Email: "a@b.com", Given: "A", Family: "B"}

	tests := []struct {
		nome   string
		token  string
		errIs  error
		errMsg string
	}{
		{"assinatura de outra chave", f.TokenSignedByOther(t, user), auth.ErrInvalidToken, ""},
		{"emissor errado", f.TokenWith(user, func(c jwt.MapClaims) { c["iss"] = "https://outro/realms/x" }), auth.ErrInvalidToken, ""},
		{"expirado", f.TokenWith(user, func(c jwt.MapClaims) { c["exp"] = 1 }), auth.ErrInvalidToken, ""},
		{"sem expiração", f.TokenWith(user, func(c jwt.MapClaims) { delete(c, "exp") }), auth.ErrInvalidToken, ""},
		{"token de ID", f.TokenWith(user, func(c jwt.MapClaims) { c["typ"] = "ID" }), auth.ErrInvalidToken, ""},
		{"token de refresh", f.TokenWith(user, func(c jwt.MapClaims) { c["typ"] = "Refresh" }), auth.ErrInvalidToken, ""},
		{"outro client", f.TokenWith(user, func(c jwt.MapClaims) { c["azp"] = "outro-app" }), auth.ErrInvalidToken, ""},
		{"lixo", "abc.def.ghi", auth.ErrInvalidToken, ""},
		{"vazio", "", auth.ErrInvalidToken, ""},
		{"sem CPF", f.TokenWith(user, func(c jwt.MapClaims) { delete(c, f.CPFClaim) }), auth.ErrMissingCPF, ""},
		{"CPF inválido", f.TokenWith(testutil.Identity{CPF: "111.111.111-11"}, nil), auth.ErrMissingCPF, ""},
		{"CPF com dígito errado", f.TokenWith(testutil.Identity{CPF: "529.982.247-26"}, nil), auth.ErrMissingCPF, ""},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			_, err := v.Verify(tt.token)
			assert.ErrorIs(t, err, tt.errIs)
		})
	}
}

func TestOIDCVerifyAceitaClientNoAudience(t *testing.T) {
	f := testutil.NewFakeOIDC(t)
	v := newVerifier(t, f)
	tok := f.TokenWith(testutil.Identity{CPF: cpfValido}, func(c jwt.MapClaims) {
		delete(c, "azp")
		c["aud"] = []string{"account", f.ClientID}
	})
	_, err := v.Verify(tok)
	assert.NoError(t, err)
}

func TestOIDCVerifyRejeitaAlgoritmoNone(t *testing.T) {
	f := testutil.NewFakeOIDC(t)
	v := newVerifier(t, f)
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"iss": f.Issuer, "azp": f.ClientID, "exp": 9999999999, f.CPFClaim: cpfValido,
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	_, err = v.Verify(none)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}
