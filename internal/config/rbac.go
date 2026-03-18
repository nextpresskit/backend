package config

import "strings"

type RBACConfig struct {
	BootstrapEnabled bool
}

func LoadRBACConfig() RBACConfig {
	v := strings.TrimSpace(strings.ToLower(GetEnv("RBAC_BOOTSTRAP_ENABLED", "false")))
	return RBACConfig{
		BootstrapEnabled: v == "1" || v == "true" || v == "yes" || v == "y" || v == "on",
	}
}

