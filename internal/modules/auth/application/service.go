package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

type RBACReader interface {
	ListRoleNamesByUserID(ctx context.Context, userID string) ([]string, error)
	ListPermissionCodesByUserID(ctx context.Context, userID string) ([]string, error)
}

type UserRelations struct {
	RoleNames       []string
	PermissionCodes []string
}

type Service struct {
	users userDomain.Repository
	tokens TokenProvider
	hasher PasswordHasher
	rbac RBACReader
}

func NewService(users userDomain.Repository, tokens TokenProvider, hasher PasswordHasher) *Service {
	return &Service{users: users, tokens: tokens, hasher: hasher}
}

func (s *Service) SetRBACReader(rbac RBACReader) {
	s.rbac = rbac
}

var (
	ErrEmailTaken   = errors.New("email already in use")
	ErrInvalidLogin = errors.New("invalid email or password")
	ErrUserNotFound = errors.New("user not found")
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
		UUID:      generateUserID(),
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

func (s *Service) Login(ctx context.Context, email, password string) (*userDomain.User, string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	u, err := s.users.FindByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if u == nil {
		return nil, "", "", ErrInvalidLogin
	}

	if err := s.hasher.CheckPassword(u.Password, password); err != nil {
		return nil, "", "", ErrInvalidLogin
	}

	access, err := s.tokens.GenerateAccessToken(fmt.Sprintf("%d", u.ID))
	if err != nil {
		return nil, "", "", err
	}
	refresh, err := s.tokens.GenerateRefreshToken(fmt.Sprintf("%d", u.ID))
	if err != nil {
		return nil, "", "", err
	}
	return u, access, refresh, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*userDomain.User, string, string, error) {
	userID, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, "", "", ErrInvalidLogin
	}

	id, err := parseUserID(userID)
	if err != nil {
		return nil, "", "", ErrInvalidLogin
	}
	u, err := s.users.FindByID(id)
	if err != nil {
		return nil, "", "", err
	}
	if u == nil {
		return nil, "", "", ErrInvalidLogin
	}

	access, err := s.tokens.GenerateAccessToken(userID)
	if err != nil {
		return nil, "", "", err
	}
	refresh, err := s.tokens.GenerateRefreshToken(userID)
	if err != nil {
		return nil, "", "", err
	}
	return u, access, refresh, nil
}

func (s *Service) Me(ctx context.Context, userID string) (*userDomain.User, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	u, err := s.users.FindByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) Relations(ctx context.Context, userID string) (UserRelations, error) {
	if s.rbac == nil {
		return UserRelations{
			RoleNames:       []string{},
			PermissionCodes: []string{},
		}, nil
	}

	roleNames, err := s.rbac.ListRoleNamesByUserID(ctx, userID)
	if err != nil {
		return UserRelations{}, err
	}
	permissionCodes, err := s.rbac.ListPermissionCodesByUserID(ctx, userID)
	if err != nil {
		return UserRelations{}, err
	}
	return UserRelations{
		RoleNames:       roleNames,
		PermissionCodes: permissionCodes,
	}, nil
}

func parseUserID(raw string) (userDomain.UserID, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || v <= 0 {
		return 0, errors.New("invalid user id")
	}
	return userDomain.UserID(v), nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	_, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return ErrInvalidLogin
	}
	return nil
}

func generateUserID() string {
	return uuid.New().String()
}
