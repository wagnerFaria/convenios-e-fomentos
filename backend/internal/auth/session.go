package auth

const (
	maxUserAgent    = 255
	tokenTypeBearer = "Bearer"
)

// TokenPair é o par de tokens devolvido no login e na renovação dos usuários finais.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"` // segundos
}

// SessionMeta descreve o cliente que abre a sessão.
type SessionMeta struct{ UserAgent, IP string }

// Trimmed limita o tamanho do user agent gravado.
func (m SessionMeta) Trimmed() SessionMeta {
	if len(m.UserAgent) > maxUserAgent {
		m.UserAgent = m.UserAgent[:maxUserAgent]
	}
	return m
}
