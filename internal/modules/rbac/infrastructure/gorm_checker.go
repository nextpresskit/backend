package infrastructure

import (
	"context"

	"gorm.io/gorm"
)

type GormPermissionChecker struct {
	db *gorm.DB
}

func NewGormPermissionChecker(db *gorm.DB) *GormPermissionChecker {
	return &GormPermissionChecker{db: db}
}

// UserHasPermission checks permission via:
// users -> user_roles -> roles -> role_permissions -> permissions.
func (c *GormPermissionChecker) UserHasPermission(ctx context.Context, userID string, permissionCode string) (bool, error) {
	var count int64

	err := c.db.WithContext(ctx).
		Table("permissions p").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Joins("JOIN roles r ON r.id = rp.role_id").
		Joins("JOIN user_roles ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Where("p.code = ?", permissionCode).
		Count(&count).
		Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

