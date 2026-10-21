// Pacote config lê e valida as variáveis de ambiente da aplicação.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config reúne todas as configurações em tempo de execução.
type Config struct {
	AppEnv   string `env:"APP_ENV" envDefault:"development"`
	HTTPPort int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://app:app@localhost:5432/app?sslmode=disable"`

	JWTAdminSecret string `env:"JWT_ADMIN_SECRET" envDefault:"dev-only-admin-secret-change-me-0000000000"`
	JWTAPISecret   string `env:"JWT_API_SECRET" envDefault:"dev-only-api-secret-change-me-000000000000"`
	APITokenSalt   string `env:"API_TOKEN_SALT" envDefault:"dev-only-api-token-salt-change-me-0000000"`

	AccessTokenTTL      time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"15m"`
	RefreshTokenMaxTTL  time.Duration `env:"REFRESH_TOKEN_MAX_TTL" envDefault:"720h"`
	RefreshTokenIdleTTL time.Duration `env:"REFRESH_TOKEN_IDLE_TTL" envDefault:"168h"`

	CORSAllowedOrigins  []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`
	RateLimitAuthPerMin int      `env:"RATE_LIMIT_AUTH_PER_MINUTE" envDefault:"10"`
	// Limite geral por IP em todas as rotas (0 desliga). Protege o banco de requisições em massa.
	RateLimitGlobalPerMin int `env:"RATE_LIMIT_GLOBAL_PER_MINUTE" envDefault:"600"`
	// Registro inicial do primeiro Super Admin por HTTP. Sem valor: habilitado fora de produção.
	InitialRegistrationEnabled *bool `env:"INITIAL_REGISTRATION_ENABLED"`
	MaxBodyBytes               int64 `env:"MAX_BODY_BYTES" envDefault:"1048576"`
	PasswordMinLength          int   `env:"PASSWORD_MIN_LENGTH" envDefault:"8"`

	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUser     string `env:"SMTP_USER"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPFrom     string `env:"SMTP_FROM" envDefault:"no-reply@localhost"`

	// Validade dos tokens de uso único.
	InviteTokenTTL time.Duration `env:"INVITE_TOKEN_TTL" envDefault:"168h"`
	ResetTokenTTL  time.Duration `env:"RESET_TOKEN_TTL" envDefault:"30m"`
	// Validade do link de confirmação de e-mail dos usuários finais.
	EmailConfirmationTTL time.Duration `env:"EMAIL_CONFIRMATION_TTL" envDefault:"72h"`

	// CIDRs de proxies reversos confiáveis; só neles o X-Forwarded-For é aceito.
	TrustedProxies []string `env:"TRUSTED_PROXIES" envSeparator:","`

	PublicURL string `env:"PUBLIC_URL" envDefault:"http://localhost:8080"`

	// Parâmetros do Argon2id (padrão recomendado pela OWASP: m=19 MiB, t=2, p=1).
	Argon2MemoryKiB  uint32 `env:"ARGON2_MEMORY_KIB" envDefault:"19456"`
	Argon2Iterations uint32 `env:"ARGON2_ITERATIONS" envDefault:"2"`
	Argon2Parallel   uint8  `env:"ARGON2_PARALLELISM" envDefault:"1"`
}

// Load lê o ambiente (carregando o .env apenas fora de produção) e valida o resultado.
func Load() (*Config, error) {
	if os := getAppEnv(); os != "production" {
		// Ausência do arquivo .env é normal.
		_ = godotenv.Load()
	}
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("lendo variáveis de ambiente: %w", err)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func getAppEnv() string {
	var c struct {
		AppEnv string `env:"APP_ENV" envDefault:"development"`
	}
	_ = env.Parse(&c)
	return c.AppEnv
}

// InitialRegistrationAllowed informa se POST /admin/auth/register-first está habilitado.
// Em produção só com INITIAL_REGISTRATION_ENABLED=true explícito; prefira a CLI (admin:create).
func (c *Config) InitialRegistrationAllowed() bool {
	if c.InitialRegistrationEnabled != nil {
		return *c.InitialRegistrationEnabled
	}
	return !c.IsProduction()
}

// IsProduction indica se a aplicação roda em produção.
func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

const (
	minSecretLen         = 32
	owaspArgon2MemoryKiB = 19456
)

// Validate verifica a consistência das configurações. Em produção, segredos
// ausentes, curtos ou iguais aos valores de desenvolvimento impedem a subida.
func (c *Config) Validate() error {
	var errs []error
	switch c.AppEnv {
	case "development", "test", "production":
	default:
		errs = append(errs, fmt.Errorf("APP_ENV inválido %q (use development, test ou production)", c.AppEnv))
	}
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		errs = append(errs, errors.New("HTTP_PORT fora do intervalo 1-65535"))
	}
	if c.AccessTokenTTL <= 0 || c.RefreshTokenMaxTTL <= 0 || c.RefreshTokenIdleTTL <= 0 {
		errs = append(errs, errors.New("as durações de token devem ser positivas"))
	}
	if c.InviteTokenTTL <= 0 || c.ResetTokenTTL <= 0 || c.EmailConfirmationTTL <= 0 {
		errs = append(errs, errors.New("INVITE_TOKEN_TTL, RESET_TOKEN_TTL e EMAIL_CONFIRMATION_TTL devem ser positivos"))
	}
	if c.PasswordMinLength < 1 {
		errs = append(errs, errors.New("PASSWORD_MIN_LENGTH deve ser >= 1"))
	}
	if c.RateLimitAuthPerMin < 1 || c.RateLimitGlobalPerMin < 0 {
		errs = append(errs, errors.New("RATE_LIMIT_AUTH_PER_MINUTE deve ser >= 1 e RATE_LIMIT_GLOBAL_PER_MINUTE >= 0"))
	}
	if c.Argon2Iterations < 1 || c.Argon2Parallel < 1 || c.Argon2MemoryKiB < 8*uint32(c.Argon2Parallel) {
		errs = append(errs, errors.New("parâmetros do Argon2 inválidos (iterações >= 1, paralelismo >= 1, memória >= 8 KiB por thread)"))
	}
	if c.IsProduction() && c.Argon2MemoryKiB < owaspArgon2MemoryKiB {
		errs = append(errs, fmt.Errorf("ARGON2_MEMORY_KIB abaixo do mínimo recomendado pela OWASP (%d) em produção", owaspArgon2MemoryKiB))
	}
	if c.IsProduction() {
		secrets := map[string]string{
			"JWT_ADMIN_SECRET": c.JWTAdminSecret,
			"JWT_API_SECRET":   c.JWTAPISecret,
			"API_TOKEN_SALT":   c.APITokenSalt,
		}
		for name, v := range secrets {
			if len(v) < minSecretLen || strings.HasPrefix(v, "dev-only-") {
				errs = append(errs, fmt.Errorf("%s ausente ou fraco: em produção use um valor aleatório com ao menos %d caracteres", name, minSecretLen))
			}
		}
		if c.JWTAdminSecret == c.JWTAPISecret {
			errs = append(errs, errors.New("JWT_ADMIN_SECRET e JWT_API_SECRET devem ser diferentes"))
		}
	}
	return errors.Join(errs...)
}
