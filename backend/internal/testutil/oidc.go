package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// FakeOIDC é um provedor OIDC de mentira: serve o JWKS por HTTP e assina access tokens no formato
// do Keycloak (MT Login), para testar a validação sem rede.
type FakeOIDC struct {
	Issuer   string
	ClientID string
	// CPFClaim é o claim que carrega o CPF nos tokens emitidos.
	CPFClaim string

	server *httptest.Server
	key    *rsa.PrivateKey
	kid    string
}

// NewFakeOIDC sobe o servidor de JWKS e o derruba ao fim do teste.
func NewFakeOIDC(t *testing.T) *FakeOIDC {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	f := &FakeOIDC{Issuer: "https://login.teste/realms/mt", ClientID: "app-teste", CPFClaim: "preferred_username", key: key, kid: "chave-1"}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{
			"kty": "RSA", "use": "sig", "alg": "RS256", "kid": f.kid,
			"n": base64.RawURLEncoding.EncodeToString(f.key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(f.key.E)).Bytes()),
		}}})
	}))
	t.Cleanup(f.server.Close)
	return f
}

// JWKSURI é a URL do JWKS deste provedor.
func (f *FakeOIDC) JWKSURI() string { return f.server.URL }

// Identity são os dados do usuário que vão no token.
type Identity struct{ CPF, Email, Given, Family string }

// Token emite um access token válido por 15 minutos para o usuário.
func (f *FakeOIDC) Token(id Identity) string { return f.TokenWith(id, nil) }

// TokenWith emite o token e deixa o teste alterar as claims antes de assinar (use nil para remover uma).
func (f *FakeOIDC) TokenWith(id Identity, mutate func(jwt.MapClaims)) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": f.Issuer, "sub": uuid.NewString(), "azp": f.ClientID, "typ": "Bearer", "aud": "account",
		"iat": now.Unix(), "exp": now.Add(15 * time.Minute).Unix(), "sid": uuid.NewString(),
		f.CPFClaim: id.CPF, "email": id.Email, "given_name": id.Given, "family_name": id.Family,
	}
	if mutate != nil {
		mutate(claims)
	}
	return f.sign(claims, f.key, f.kid)
}

// TokenSignedByOther emite um token com assinatura de outra chave (não publicada no JWKS).
func (f *FakeOIDC) TokenSignedByOther(t *testing.T, id Identity) string {
	t.Helper()
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	now := time.Now()
	return f.sign(jwt.MapClaims{
		"iss": f.Issuer, "azp": f.ClientID, "typ": "Bearer", "exp": now.Add(time.Minute).Unix(), f.CPFClaim: id.CPF,
	}, other, f.kid)
}

func (f *FakeOIDC) sign(claims jwt.MapClaims, key *rsa.PrivateKey, kid string) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		panic(err)
	}
	return s
}
