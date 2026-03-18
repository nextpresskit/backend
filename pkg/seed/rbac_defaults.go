package seed

import "gorm.io/gorm"

// Deterministic UUIDs so environments stay consistent.
const (
	RoleAdminID = "00000000-0000-0000-0000-000000000001"

	PermissionAdminPingID = "00000000-0000-0000-0000-000000000101"
	PermissionRBACManageID = "00000000-0000-0000-0000-000000000102"
	PermissionPostsReadID  = "00000000-0000-0000-0000-000000000201"
	PermissionPostsWriteID = "00000000-0000-0000-0000-000000000202"
	PermissionPagesReadID  = "00000000-0000-0000-0000-000000000203"
	PermissionPagesWriteID = "00000000-0000-0000-0000-000000000204"
	PermissionCategoriesReadID  = "00000000-0000-0000-0000-000000000205"
	PermissionCategoriesWriteID = "00000000-0000-0000-0000-000000000206"
	PermissionTagsReadID        = "00000000-0000-0000-0000-000000000207"
	PermissionTagsWriteID       = "00000000-0000-0000-0000-000000000208"
	PermissionMediaReadID       = "00000000-0000-0000-0000-000000000209"
	PermissionMediaWriteID      = "00000000-0000-0000-0000-000000000210"
	PermissionMenusReadID       = "00000000-0000-0000-0000-000000000211"
	PermissionMenusWriteID      = "00000000-0000-0000-0000-000000000212"
)

func SeedRBACDefaults(db *gorm.DB) error {
	// roles
	if err := db.Exec(
		`INSERT INTO roles (id, name) VALUES (?, ?) ON CONFLICT (name) DO NOTHING`,
		RoleAdminID, "admin",
	).Error; err != nil {
		return err
	}

	// permissions
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionAdminPingID, "admin:ping",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionRBACManageID, "rbac:manage",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionPostsReadID, "posts:read",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionPostsWriteID, "posts:write",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionPagesReadID, "pages:read",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionPagesWriteID, "pages:write",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionCategoriesReadID, "categories:read",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionCategoriesWriteID, "categories:write",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionTagsReadID, "tags:read",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionTagsWriteID, "tags:write",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionMediaReadID, "media:read",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionMediaWriteID, "media:write",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionMenusReadID, "menus:read",
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO permissions (id, code) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
		PermissionMenusWriteID, "menus:write",
	).Error; err != nil {
		return err
	}

	// role_permissions links
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionAdminPingID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionRBACManageID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionPostsReadID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionPostsWriteID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionPagesReadID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionPagesWriteID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionCategoriesReadID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionCategoriesWriteID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionTagsReadID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionTagsWriteID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionMediaReadID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionMediaWriteID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionMenusReadID,
	).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?) ON CONFLICT DO NOTHING`,
		RoleAdminID, PermissionMenusWriteID,
	).Error; err != nil {
		return err
	}

	return nil
}

