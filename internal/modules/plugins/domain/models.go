package domain

import "time"

type PluginID string

type Plugin struct {
	ID        PluginID
	Name      string
	Slug      string
	Enabled   bool
	Version   string
	Config    map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

