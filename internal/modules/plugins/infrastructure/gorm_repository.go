package infrastructure

import (
	"context"
	"encoding/json"

	pluginsDomain "github.com/nextpresskit/backend/internal/modules/plugins/domain"
	pluginp "github.com/nextpresskit/backend/internal/modules/plugins/persistence"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, plugin *pluginsDomain.Plugin) error {
	cfgBytes, err := marshalConfig(plugin.Config)
	if err != nil {
		return err
	}

	m := pluginp.Plugin{
		ID:        string(plugin.ID),
		Name:      plugin.Name,
		Slug:      plugin.Slug,
		Enabled:   plugin.Enabled,
		Version:   plugin.Version,
		ConfigRaw: cfgBytes,
		CreatedAt: plugin.CreatedAt,
		UpdatedAt: plugin.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}

	*plugin = *toDomain(&m)
	return nil
}

func (r *GormRepository) FindByID(ctx context.Context, id pluginsDomain.PluginID) (*pluginsDomain.Plugin, error) {
	var row pluginp.Plugin
	if err := r.db.WithContext(ctx).
		Where("id = ?", string(id)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return toDomain(&row), nil
}

func (r *GormRepository) FindBySlug(ctx context.Context, slug string) (*pluginsDomain.Plugin, error) {
	var row pluginp.Plugin
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return toDomain(&row), nil
}

func (r *GormRepository) List(ctx context.Context) ([]pluginsDomain.Plugin, error) {
	var rows []pluginp.Plugin
	if err := r.db.WithContext(ctx).
		Model(&pluginp.Plugin{}).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]pluginsDomain.Plugin, 0, len(rows))
	for i := range rows {
		out = append(out, *toDomain(&rows[i]))
	}
	return out, nil
}

func (r *GormRepository) ListEnabled(ctx context.Context) ([]pluginsDomain.Plugin, error) {
	var rows []pluginp.Plugin
	if err := r.db.WithContext(ctx).
		Model(&pluginp.Plugin{}).
		Where("enabled = ?", true).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]pluginsDomain.Plugin, 0, len(rows))
	for i := range rows {
		out = append(out, *toDomain(&rows[i]))
	}
	return out, nil
}

func (r *GormRepository) Update(ctx context.Context, plugin *pluginsDomain.Plugin) error {
	cfgBytes, err := marshalConfig(plugin.Config)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"name":       plugin.Name,
		"slug":       plugin.Slug,
		"enabled":    plugin.Enabled,
		"version":    plugin.Version,
		"config":     cfgBytes,
		"updated_at": plugin.UpdatedAt,
	}

	return r.db.WithContext(ctx).
		Model(&pluginp.Plugin{}).
		Where("id = ?", string(plugin.ID)).
		Updates(updates).Error
}

func toDomain(m *pluginp.Plugin) *pluginsDomain.Plugin {
	cfg := map[string]any{}
	if len(m.ConfigRaw) > 0 {
		_ = json.Unmarshal(m.ConfigRaw, &cfg)
	}

	return &pluginsDomain.Plugin{
		ID:        pluginsDomain.PluginID(m.ID),
		Name:      m.Name,
		Slug:      m.Slug,
		Enabled:   m.Enabled,
		Version:   m.Version,
		Config:    cfg,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func marshalConfig(cfg map[string]any) ([]byte, error) {
	if cfg == nil {
		cfg = map[string]any{}
	}
	return json.Marshal(cfg)
}

