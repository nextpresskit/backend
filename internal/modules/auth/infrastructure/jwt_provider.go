package infrastructure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func NewJWTProvider(secret string, accessTTL, refreshTTL time.Duration) *JWTProvider {
	return &JWTProvider{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (p *JWTProvider) GenerateAccessToken(userID string) (string, error) {
	return p.generate(userID, p.accessTTL)
}

func (p *JWTProvider) GenerateRefreshToken(userID string) (string, error) {
	return p.generate(userID, p.refreshTTL)
}

func (p *JWTProvider) generate(userID string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString(p.secret)
}

func (p *JWTProvider) ParseAccessToken(tokenStr string) (string, error) {
	return p.parse(tokenStr)
}

func (p *JWTProvider) ParseRefreshToken(tokenStr string) (string, error) {
	return p.parse(tokenStr)
}

func (p *JWTProvider) parse(tokenStr string) (string, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return p.secret, nil
	})
	if err != nil {
		return "", err
	}
	if c, ok := t.Claims.(*claims); ok && t.Valid {
		return c.UserID, nil
	}
	return "", jwt.ErrTokenInvalidClaims
}
