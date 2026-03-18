package application

import (
	"context"
	"errors"
	"strings"

	userDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
	"github.com/google/uuid"
)

type TokenProvider interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ParseAccessToken(token string) (string, error)
	ParseRefreshToken(token string) (string, error)
}

type PasswordHasher interface {
	HashPassword(plain string) (string, error)
	CheckPassword(hash, plain string) error
}

type Service struct {
	users  userDomain.Repository
	tokens TokenProvider
	hasher PasswordHasher
}

func NewService(users userDomain.Repository, tokens TokenProvider, hasher PasswordHasher) *Service {
	return &Service{users: users, tokens: tokens, hasher: hasher}
}

var (
	ErrEmailTaken   = errors.New("email already in use")
	ErrInvalidLogin = errors.New("invalid email or password")
)

func (s *Service) Register(ctx context.Context, firstName, lastName, email, password string) (*userDomain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	existing, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &userDomain.User{
		ID:        userDomain.UserID(generateUserID()),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  hash,
		Active:    true,
	}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	email = strings.TrimSpace(strings.ToLower(email))

	u, err := s.users.FindByEmail(email)
	if err != nil {
		return "", "", err
	}
	if u == nil {
		return "", "", ErrInvalidLogin
	}

	if err := s.hasher.CheckPassword(u.Password, password); err != nil {
		return "", "", ErrInvalidLogin
	}

	access, err = s.tokens.GenerateAccessToken(string(u.ID))
	if err != nil {
		return "", "", err
	}
	refresh, err = s.tokens.GenerateRefreshToken(string(u.ID))
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (access, refresh string, err error) {
	userID, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidLogin
	}

	access, err = s.tokens.GenerateAccessToken(userID)
	if err != nil {
		return "", "", err
	}
	refresh, err = s.tokens.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func generateUserID() string {
	return uuid.New().String()
}
