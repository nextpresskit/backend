package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	rbacDomain "github.com/nextpresskit/backend/internal/modules/rbac/domain"
)

var (
	ErrInvalidNameOrCode = errors.New("invalid_name_or_code")
	ErrAlreadyExists     = errors.New("already_exists")
)

type Service struct {
	repo rbacDomain.Repository
}

func NewService(repo rbacDomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRole(ctx context.Context, name string) (*rbacDomain.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidNameOrCode
	}

	now := time.Now().UTC()
	role := &rbacDomain.Role{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return role, nil
}

func (s *Service) ListRoles(ctx context.Context) ([]rbacDomain.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) CreatePermission(ctx context.Context, code string) (*rbacDomain.Permission, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrInvalidNameOrCode
	}

	now := time.Now().UTC()
	perm := &rbacDomain.Permission{
		ID:        uuid.NewString(),
		Code:      code,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreatePermission(ctx, perm); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return perm, nil
}

func (s *Service) ListPermissions(ctx context.Context) ([]rbacDomain.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *Service) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	userID = strings.TrimSpace(userID)
	roleID = strings.TrimSpace(roleID)
	if userID == "" || roleID == "" {
		return ErrInvalidNameOrCode
	}
	return s.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (s *Service) GrantPermissionToRole(ctx context.Context, roleID string, permissionID string) error {
	roleID = strings.TrimSpace(roleID)
	permissionID = strings.TrimSpace(permissionID)
	if roleID == "" || permissionID == "" {
		return ErrInvalidNameOrCode
	}
	return s.repo.GrantPermissionToRole(ctx, roleID, permissionID)
}

