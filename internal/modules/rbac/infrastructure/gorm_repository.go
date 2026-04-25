package infrastructure

import (
	"context"
	"strconv"
	"strings"
	"time"

	rbacDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/rbac/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

type gormRole struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (gormRole) TableName() string { return "roles" }

type gormPermission struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Code      string    `gorm:"column:code"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (gormPermission) TableName() string { return "permissions" }

type gormUserRole struct {
	UserID int64  `gorm:"column:user_id;primaryKey"`
	RoleID string `gorm:"column:role_id;type:uuid;primaryKey"`
}

func (gormUserRole) TableName() string { return "user_roles" }

type gormRolePermission struct {
	RoleID       string `gorm:"column:role_id;type:uuid;primaryKey"`
	PermissionID string `gorm:"column:permission_id;type:uuid;primaryKey"`
}

func (gormRolePermission) TableName() string { return "role_permissions" }

func (r *GormRepository) CreateRole(ctx context.Context, role *rbacDomain.Role) error {
	m := gormRole{
		ID:        role.ID,
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormRepository) ListRoles(ctx context.Context) ([]rbacDomain.Role, error) {
	var rows []gormRole
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]rbacDomain.Role, 0, len(rows))
	for _, row := range rows {
		out = append(out, rbacDomain.Role{
			ID:        row.ID,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) CreatePermission(ctx context.Context, perm *rbacDomain.Permission) error {
	m := gormPermission{
		ID:        perm.ID,
		Code:      perm.Code,
		CreatedAt: perm.CreatedAt,
		UpdatedAt: perm.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *GormRepository) ListPermissions(ctx context.Context) ([]rbacDomain.Permission, error) {
	var rows []gormPermission
	if err := r.db.WithContext(ctx).Order("code ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]rbacDomain.Permission, 0, len(rows))
	for _, row := range rows {
		out = append(out, rbacDomain.Permission{
			ID:        row.ID,
			Code:      row.Code,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	uid, err := strconv.ParseInt(strings.TrimSpace(userID), 10, 64)
	if err != nil {
		return err
	}
	m := gormUserRole{UserID: uid, RoleID: roleID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&m).
		Error
}

func (r *GormRepository) GrantPermissionToRole(ctx context.Context, roleID string, permissionID string) error {
	m := gormRolePermission{RoleID: roleID, PermissionID: permissionID}
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

