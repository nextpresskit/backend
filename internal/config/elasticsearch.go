package config

import (
	"os"
	"strings"
)

// DefaultElasticsearchIndexPrefix is used when ELASTICSEARCH_INDEX_PREFIX is unset or empty.
const DefaultElasticsearchIndexPrefix = "nextpresskit"

// ElasticsearchConfig toggles Elasticsearch client, indexing, and search.
type ElasticsearchConfig struct {
	Enabled bool

	// Addresses is one or more node URLs (comma-separated), e.g. "https://localhost:9200".
	Addresses []string

	APIKey   string
	Username string
	Password string

	IndexPrefix string

	// InsecureSkipVerify disables TLS certificate verification (local/dev only).
	InsecureSkipVerify bool

	// AutoCreateIndex creates the posts index on startup when missing (local/dev only).
	AutoCreateIndex bool
}

// LoadElasticsearchConfig reads ELASTICSEARCH_* environment variables.
func LoadElasticsearchConfig(appEnv string) ElasticsearchConfig {
	addrsRaw := GetEnv("ELASTICSEARCH_URLS", "")
	var addrs []string
	for _, p := range strings.Split(addrsRaw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			addrs = append(addrs, p)
		}
	}
	env := strings.ToLower(strings.TrimSpace(appEnv))
	autoCreateDefault := "false"
	if env == "local" || env == "dev" {
		autoCreateDefault = "true"
	}
	var autoCreateIndex bool
	if _, ok := os.LookupEnv("ELASTICSEARCH_AUTO_CREATE_INDEX"); ok {
		autoCreateIndex = parseBool(strings.TrimSpace(os.Getenv("ELASTICSEARCH_AUTO_CREATE_INDEX")))
	} else {
		autoCreateIndex = parseBool(autoCreateDefault)
	}
	return ElasticsearchConfig{
		Enabled:            parseBool(GetEnv("ELASTICSEARCH_ENABLED", "false")),
		Addresses:          addrs,
		APIKey:             GetEnv("ELASTICSEARCH_API_KEY", ""),
		Username:           GetEnv("ELASTICSEARCH_USERNAME", ""),
		Password:           GetEnv("ELASTICSEARCH_PASSWORD", ""),
		IndexPrefix:        GetEnv("ELASTICSEARCH_INDEX_PREFIX", DefaultElasticsearchIndexPrefix),
		InsecureSkipVerify: parseBool(GetEnv("ELASTICSEARCH_INSECURE_SKIP_VERIFY", "false")),
		AutoCreateIndex:    autoCreateIndex,
	}
}
