package application

import (
	userDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
)

type TokenProvider interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ParseAccessToken(token string) (string, error)
}

type Service struct {
	users  userDomain.Repository
	tokens TokenProvider
	hasher PasswordHasher
}

type PasswordHasher interface {
	HashPassword(plain string) (string, error)
	CheckPassword(hash, plain string) error
}

func NewService(users userDomain.Repository, tokens TokenProvider, hasher PasswordHasher) *Service {
	return &Service{users: users, tokens: tokens, hasher: hasher}
}
