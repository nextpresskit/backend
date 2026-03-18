package domain

import "context"

type Repository interface {
	// Roles / permissions
	CreateRole(ctx context.Context, role *Role) error
	ListRoles(ctx context.Context) ([]Role, error)
	CreatePermission(ctx context.Context, perm *Permission) error
	ListPermissions(ctx context.Context) ([]Permission, error)

	// Assignments
	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	GrantPermissionToRole(ctx context.Context, roleID string, permissionID string) error
}

