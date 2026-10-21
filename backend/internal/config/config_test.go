package config

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProducao(t *testing.T) {
	forte := func() *Config {
		return &Config{
			AppEnv: "production", HTTPPort: 8080, PasswordMinLength: 8,
			AccessTokenTTL: 1, RefreshTokenMaxTTL: 1, RefreshTokenIdleTTL: 1,
			InviteTokenTTL: 1, ResetTokenTTL: 1, EmailConfirmationTTL: 1, RateLimitAuthPerMin: 10,
			Argon2MemoryKiB: 19456, Argon2Iterations: 2, Argon2Parallel: 1,
			JWTAdminSecret: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			JWTAPISecret:   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			APITokenSalt:   "cccccccccccccccccccccccccccccccccccc",
		}
	}
	tests := []struct {
		nome    string
		mutar   func(*Config)
		quebrou bool
	}{
		{"segredos fortes", func(*Config) {}, false},
		{"segredo curto", func(c *Config) { c.JWTAdminSecret = "curto" }, true},
		{"segredo de desenvolvimento", func(c *Config) { c.APITokenSalt = "dev-only-api-token-salt-change-me-0000000" }, true},
		{"segredos jwt iguais", func(c *Config) { c.JWTAPISecret = c.JWTAdminSecret }, true},
		{"ambiente inválido", func(c *Config) { c.AppEnv = "x" }, true},
		{"argon2 fraco em produção", func(c *Config) { c.Argon2MemoryKiB = 1024 }, true},
		{"argon2 sem iterações", func(c *Config) { c.Argon2Iterations = 0 }, true},
		{"limite de login zerado", func(c *Config) { c.RateLimitAuthPerMin = 0 }, true},
		{"validade de token zerada", func(c *Config) { c.EmailConfirmationTTL = 0 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			c := forte()
			tt.mutar(c)
			if tt.quebrou {
				assert.Error(t, c.Validate())
			} else {
				assert.NoError(t, c.Validate())
			}
		})
	}
}

func TestSegredosDeDesenvolvimentoSoForaDeProducao(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	c, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, 8080, c.HTTPPort)
}

func TestRegistroInicialPorAmbiente(t *testing.T) {
	yes, no := true, false
	tests := []struct {
		nome string
		c    Config
		want bool
	}{
		{"desenvolvimento: habilitado por padrão", Config{AppEnv: "development"}, true},
		{"produção: desabilitado por padrão", Config{AppEnv: "production"}, false},
		{"produção: habilitado explicitamente", Config{AppEnv: "production", InitialRegistrationEnabled: &yes}, true},
		{"desenvolvimento: desabilitado explicitamente", Config{AppEnv: "development", InitialRegistrationEnabled: &no}, false},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) { assert.Equal(t, tt.want, tt.c.InitialRegistrationAllowed()) })
	}
}

// Toda variável lida pela aplicação precisa estar documentada no .env.example, e vice-versa.
func TestEnvExampleDocumentaTodasAsVariaveis(t *testing.T) {
	raw, err := os.ReadFile("../../.env.example")
	require.NoError(t, err)
	documented := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
		if name, _, ok := strings.Cut(line, "="); ok && name == strings.ToUpper(name) && !strings.ContainsAny(name, " ") && name != "" {
			documented[name] = true
		}
	}
	read := map[string]bool{"DB_PORT": true} // usada só pelo docker-compose.yml
	typ := reflect.TypeOf(Config{})
	for i := range typ.NumField() {
		tag, _, _ := strings.Cut(typ.Field(i).Tag.Get("env"), ",")
		if tag != "" {
			read[tag] = true
		}
	}
	for name := range read {
		assert.True(t, documented[name], "%s não está no .env.example", name)
	}
	for name := range documented {
		assert.True(t, read[name], "%s está no .env.example mas não é lida pela aplicação", name)
	}
}
