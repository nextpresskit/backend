package infrastructure

import "github.com/golang-jwt/jwt/v5"

type JWTProvider struct {
	secret []byte
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secret: []byte(secret)}
}

// Methods will be implemented when wiring real JWT logic in the next step.
func (p *JWTProvider) GenerateAccessToken(userID string) (string, error)  { return "", nil }
func (p *JWTProvider) GenerateRefreshToken(userID string) (string, error) { return "", nil }
func (p *JWTProvider) ParseAccessToken(token string) (string, error)      { return "", nil }

var _ jwt.Claims
