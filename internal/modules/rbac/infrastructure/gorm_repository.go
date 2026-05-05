package infrastructure

import (
	"context"
	"strconv"
	"strings"

	rbacDomain "github.com/nextpresskit/backend/internal/modules/rbac/domain"
	rbacp "github.com/nextpresskit/backend/internal/modules/rbac/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateRole(ctx context.Context, role *rbacDomain.Role) error {
	m := rbacp.Role{
		UUID:      role.UUID,
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormRepository) ListRoles(ctx context.Context) ([]rbacDomain.Role, error) {
	var rows []rbacp.Role
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]rbacDomain.Role, 0, len(rows))
	for _, row := range rows {
		out = append(out, rbacDomain.Role{
			ID:        row.ID,
			UUID:      row.UUID,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) CreatePermission(ctx context.Context, perm *rbacDomain.Permission) error {
	m := rbacp.Permission{
		UUID:      perm.UUID,
		Code:      perm.Code,
		CreatedAt: perm.CreatedAt,
		UpdatedAt: perm.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormRepository) ListPermissions(ctx context.Context) ([]rbacDomain.Permission, error) {
	var rows []rbacp.Permission
	if err := r.db.WithContext(ctx).Order("code ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]rbacDomain.Permission, 0, len(rows))
	for _, row := range rows {
		out = append(out, rbacDomain.Permission{
			ID:        row.ID,
			UUID:      row.UUID,
			Code:      row.Code,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) AssignRoleToUser(ctx context.Context, userID string, roleUUID string) error {
	uid, err := strconv.ParseInt(strings.TrimSpace(userID), 10, 64)
	if err != nil {
		return err
	}
	var role rbacp.Role
	if err := r.db.WithContext(ctx).Where("uuid = ?", strings.TrimSpace(roleUUID)).First(&role).Error; err != nil {
		return err
	}
	m := rbacp.UserRole{UserID: uid, RoleID: role.ID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&m).
		Error
}

func (r *GormRepository) GrantPermissionToRole(ctx context.Context, roleUUID string, permissionUUID string) error {
	var role rbacp.Role
	if err := r.db.WithContext(ctx).Where("uuid = ?", strings.TrimSpace(roleUUID)).First(&role).Error; err != nil {
		return err
	}
	var perm rbacp.Permission
	if err := r.db.WithContext(ctx).Where("uuid = ?", strings.TrimSpace(permissionUUID)).First(&perm).Error; err != nil {
		return err
	}
	m := rbacp.RolePermission{RoleID: role.ID, PermissionID: perm.ID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&m).
		Error
}

func (r *GormRepository) ListRoleNamesByUserID(ctx context.Context, userID string) ([]string, error) {
	var names []string
	uid, err := strconv.ParseInt(strings.TrimSpace(userID), 10, 64)
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).
		Table("roles AS r").
		Select("r.name").
		Joins("JOIN user_roles AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", uid).
		Order("r.name ASC").
		Pluck("r.name", &names).Error
	if err != nil {
		return nil, err
	}
	return names, nil
}

func (r *GormRepository) ListPermissionCodesByUserID(ctx context.Context, userID string) ([]string, error) {
	var codes []string
	uid, err := strconv.ParseInt(strings.TrimSpace(userID), 10, 64)
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).
		Table("permissions AS p").
		Select("DISTINCT p.code").
		Joins("JOIN role_permissions AS rp ON rp.permission_id = p.id").
		Joins("JOIN user_roles AS ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", uid).
		Order("p.code ASC").
		Pluck("p.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}
