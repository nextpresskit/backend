package domain

import "context"

// PermissionChecker answers the question "does user X have permission Y?"
// without exposing transport (HTTP) or persistence (GORM) details.
type PermissionChecker interface {
	UserHasPermission(ctx context.Context, userID string, permissionCode string) (bool, error)
}

