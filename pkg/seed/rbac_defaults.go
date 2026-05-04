package seed

import (
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
	PermissionPluginsManageID   = "00000000-0000-0000-0000-000000000213"
)

func SeedRBACDefaults(db *gorm.DB) error {
	roles := []rbacp.Role{
		{ID: RoleAdminID, Name: "admin"},
	}
	perms := []rbacp.Permission{
		{ID: PermissionAdminPingID, Code: "admin:ping"},
		{ID: PermissionRBACManageID, Code: "rbac:manage"},
		{ID: PermissionPostsReadID, Code: "posts:read"},
		{ID: PermissionPostsWriteID, Code: "posts:write"},
		{ID: PermissionPagesReadID, Code: "pages:read"},
		{ID: PermissionPagesWriteID, Code: "pages:write"},
		{ID: PermissionCategoriesReadID, Code: "categories:read"},
		{ID: PermissionCategoriesWriteID, Code: "categories:write"},
		{ID: PermissionTagsReadID, Code: "tags:read"},
		{ID: PermissionTagsWriteID, Code: "tags:write"},
		{ID: PermissionMediaReadID, Code: "media:read"},
		{ID: PermissionMediaWriteID, Code: "media:write"},
		{ID: PermissionPluginsManageID, Code: "plugins:manage"},
	}
	for i := range roles {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&roles[i]).Error; err != nil {
			return err
		}
	}
	for i := range perms {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&perms[i]).Error; err != nil {
			return err
		}
	}
	links := []rbacp.RolePermission{
		{RoleID: RoleAdminID, PermissionID: PermissionAdminPingID},
		{RoleID: RoleAdminID, PermissionID: PermissionRBACManageID},
		{RoleID: RoleAdminID, PermissionID: PermissionPostsReadID},
		{RoleID: RoleAdminID, PermissionID: PermissionPostsWriteID},
		{RoleID: RoleAdminID, PermissionID: PermissionPagesReadID},
		{RoleID: RoleAdminID, PermissionID: PermissionPagesWriteID},
		{RoleID: RoleAdminID, PermissionID: PermissionCategoriesReadID},
		{RoleID: RoleAdminID, PermissionID: PermissionCategoriesWriteID},
		{RoleID: RoleAdminID, PermissionID: PermissionTagsReadID},
		{RoleID: RoleAdminID, PermissionID: PermissionTagsWriteID},
		{RoleID: RoleAdminID, PermissionID: PermissionMediaReadID},
		{RoleID: RoleAdminID, PermissionID: PermissionMediaWriteID},
		{RoleID: RoleAdminID, PermissionID: PermissionPluginsManageID},
	}
	for i := range links {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&links[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
