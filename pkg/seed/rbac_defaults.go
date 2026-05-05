package seed

import (
	"fmt"

	rbacp "github.com/nextpresskit/backend/internal/modules/rbac/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Deterministic UUIDs so environments stay consistent.
const (
	RoleAdminID = "00000000-0000-0000-0000-000000000001"

	PermissionAdminPingID       = "00000000-0000-0000-0000-000000000101"
	PermissionRBACManageID      = "00000000-0000-0000-0000-000000000102"
	PermissionPostsReadID       = "00000000-0000-0000-0000-000000000201"
	PermissionPostsWriteID      = "00000000-0000-0000-0000-000000000202"
	PermissionPagesReadID       = "00000000-0000-0000-0000-000000000203"
	PermissionPagesWriteID      = "00000000-0000-0000-0000-000000000204"
	PermissionCategoriesReadID  = "00000000-0000-0000-0000-000000000205"
	PermissionCategoriesWriteID = "00000000-0000-0000-0000-000000000206"
	PermissionTagsReadID        = "00000000-0000-0000-0000-000000000207"
	PermissionTagsWriteID       = "00000000-0000-0000-0000-000000000208"
	PermissionMediaReadID       = "00000000-0000-0000-0000-000000000209"
	PermissionMediaWriteID      = "00000000-0000-0000-0000-000000000210"
)

func knownPermissionID(code string) (string, bool) {
	m := map[string]string{
		"admin:ping":         PermissionAdminPingID,
		"rbac:manage":        PermissionRBACManageID,
		"posts:read":         PermissionPostsReadID,
		"posts:write":        PermissionPostsWriteID,
		"pages:read":         PermissionPagesReadID,
		"pages:write":        PermissionPagesWriteID,
		"categories:read":    PermissionCategoriesReadID,
		"categories:write":   PermissionCategoriesWriteID,
		"tags:read":          PermissionTagsReadID,
		"tags:write":         PermissionTagsWriteID,
		"media:read":         PermissionMediaReadID,
		"media:write":        PermissionMediaWriteID,
	}
	id, ok := m[code]
	return id, ok
}

// SeedRBACDefaults upserts the admin role, permission rows for known codes, and admin role links.
func SeedRBACDefaults(db *gorm.DB, permissionCodes []string) error {
	admin := rbacp.Role{UUID: RoleAdminID, Name: "admin"}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoNothing: true,
	}).Create(&admin).Error; err != nil {
		return err
	}
	var roleRow rbacp.Role
	if err := db.Where("uuid = ?", RoleAdminID).First(&roleRow).Error; err != nil {
		return fmt.Errorf("load admin role: %w", err)
	}
	for _, code := range permissionCodes {
		permUUID, ok := knownPermissionID(code)
		if !ok {
			continue
		}
		p := rbacp.Permission{UUID: permUUID, Code: code}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoUpdates: clause.AssignmentColumns([]string{"uuid", "updated_at"}),
		}).Create(&p).Error; err != nil {
			return err
		}
		var permRow rbacp.Permission
		if err := db.Where("code = ?", code).First(&permRow).Error; err != nil {
			return err
		}
		link := rbacp.RolePermission{RoleID: roleRow.ID, PermissionID: permRow.ID}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
			return err
		}
	}
	return nil
}
