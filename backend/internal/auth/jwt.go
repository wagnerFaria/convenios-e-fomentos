package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Audiências dos JWTs: um token de um domínio é rejeitado no outro.
const (
	AudienceAdmin = "admin"
	AudienceAPI   = "api"
)

const issuer = "convenios-e-fomentos"

// ErrInvalidToken indica JWT malformado, expirado, de outra audiência ou assinatura inválida.
var ErrInvalidToken = errors.New("token inválido")

// Claims são as claims dos access tokens.
type Claims struct {
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

// SubjectID devolve o sujeito (admin ou usuário final) do token.
func (c Claims) SubjectID() (uuid.UUID, error) { return uuid.Parse(c.Subject) }

// JWTIssuer emite e valida access tokens HS256 de uma audiência.
type JWTIssuer struct {
	secret   []byte
	audience string
	ttl      time.Duration
	now      func() time.Time
}

// NewJWTIssuer cria um emissor para a audiência informada.
func NewJWTIssuer(secret, audience string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), audience: audience, ttl: ttl, now: time.Now}
}

// TTL é a validade dos tokens emitidos.
func (j *JWTIssuer) TTL() time.Duration { return j.ttl }

// Issue assina um access token para o sujeito e a sessão.
func (j *JWTIssuer) Issue(subject, session uuid.UUID) (string, error) {
	now := j.now()
	claims := Claims{
		SessionID: session,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject.String(),
			Audience:  jwt.ClaimStrings{j.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("assinando JWT: %w", err)
	}
	return s, nil
}

// Parse valida assinatura, algoritmo, emissor, audiência e expiração.
func (j *JWTIssuer) Parse(token string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithAudience(j.audience),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(j.now),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	return &c, nil
}
