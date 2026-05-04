package application

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	rbacDomain "github.com/nextpresskit/backend/internal/modules/rbac/domain"
)

type repoStub struct {
	createRoleErr       error
	createPermissionErr error
	roles               []rbacDomain.Role
	permissions         []rbacDomain.Permission
	assignCalled        bool
	grantCalled         bool
}

func (s *repoStub) CreateRole(_ context.Context, role *rbacDomain.Role) error {
	if s.createRoleErr != nil {
		return s.createRoleErr
	}
	s.roles = append(s.roles, *role)
	return nil
}
func (s *repoStub) ListRoles(_ context.Context) ([]rbacDomain.Role, error) { return s.roles, nil }
func (s *repoStub) CreatePermission(_ context.Context, perm *rbacDomain.Permission) error {
	if s.createPermissionErr != nil {
		return s.createPermissionErr
	}
	s.permissions = append(s.permissions, *perm)
	return nil
}
func (s *repoStub) ListPermissions(_ context.Context) ([]rbacDomain.Permission, error) {
	return s.permissions, nil
}
func (s *repoStub) AssignRoleToUser(_ context.Context, _, _ string) error {
	s.assignCalled = true
	return nil
}
func (s *repoStub) GrantPermissionToRole(_ context.Context, _, _ string) error {
	s.grantCalled = true
	return nil
}
func (s *repoStub) ListRoleNamesByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (s *repoStub) ListPermissionCodesByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestCreateRole_Valid(t *testing.T) {
	repo := &repoStub{}
	svc := NewService(repo)

	role, err := svc.CreateRole(context.Background(), "  editor ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role == nil || role.Name != "editor" {
		t.Fatalf("expected trimmed role name, got %#v", role)
	}
}

func TestCreateRole_AlreadyExists(t *testing.T) {
	repo := &repoStub{createRoleErr: gorm.ErrDuplicatedKey}
	svc := NewService(repo)

	_, err := svc.CreateRole(context.Background(), "editor")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestCreatePermission_Invalid(t *testing.T) {
	svc := NewService(&repoStub{})
	_, err := svc.CreatePermission(context.Background(), " ")
	if !errors.Is(err, ErrInvalidNameOrCode) {
		t.Fatalf("expected ErrInvalidNameOrCode, got %v", err)
	}
}

func TestAssignRoleToUser_InvalidInput(t *testing.T) {
	svc := NewService(&repoStub{})
	err := svc.AssignRoleToUser(context.Background(), " ", "role-1")
	if !errors.Is(err, ErrInvalidNameOrCode) {
		t.Fatalf("expected ErrInvalidNameOrCode, got %v", err)
	}
}

func TestGrantPermissionToRole_CallsRepo(t *testing.T) {
	repo := &repoStub{}
	svc := NewService(repo)

	if err := svc.GrantPermissionToRole(context.Background(), "role-1", "perm-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.grantCalled {
		t.Fatal("expected repository GrantPermissionToRole to be called")
	}
}

