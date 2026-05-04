package config

import (
	"testing"
)

func TestLoadGraphQLConfig(t *testing.T) {
	t.Setenv("GRAPHQL_ENABLED", "true")
	t.Setenv("GRAPHQL_PATH", "/custom/graphql")
	t.Setenv("GRAPHQL_PLAYGROUND_ENABLED", "1")
	c := LoadGraphQLConfig()
	if !c.Enabled {
		t.Fatal("expected Enabled true")
	}
	if c.Path != "/custom/graphql" {
		t.Fatalf("path: got %q", c.Path)
	}
	if !c.PlaygroundEnabled {
		t.Fatal("expected PlaygroundEnabled true")
	}
}

func TestLoadGraphQLConfig_DefaultPathFromAPIBasePath(t *testing.T) {
	t.Setenv("GRAPHQL_PATH", "")
	t.Setenv("API_BASE_PATH", "/v1/")
	c := LoadGraphQLConfig()
	if c.Path != "/v1/graphql" {
		t.Fatalf("path: got %q", c.Path)
	}
}

func TestLoadAppConfig_NormalizesAPIBasePath(t *testing.T) {
	t.Setenv("API_BASE_PATH", "v2/")
	c := LoadAppConfig()
	if c.APIBasePath != "/v2" {
		t.Fatalf("APIBasePath: got %q", c.APIBasePath)
	}
}

func TestLoadElasticsearchConfig_Defaults(t *testing.T) {
	t.Setenv("ELASTICSEARCH_ENABLED", "false")
	t.Setenv("ELASTICSEARCH_URLS", "")
	t.Setenv("ELASTICSEARCH_INDEX_PREFIX", "nextpresskit")
	t.Setenv("ELASTICSEARCH_API_KEY", "")
	t.Setenv("ELASTICSEARCH_USERNAME", "")
	t.Setenv("ELASTICSEARCH_PASSWORD", "")
	t.Setenv("ELASTICSEARCH_INSECURE_SKIP_VERIFY", "false")
	t.Setenv("ELASTICSEARCH_AUTO_CREATE_INDEX", "false")
	c := LoadElasticsearchConfig("production")
	if c.Enabled {
		t.Fatal("expected disabled")
	}
	if c.IndexPrefix != "nextpresskit" {
		t.Fatalf("IndexPrefix: %q", c.IndexPrefix)
	}
	if c.AutoCreateIndex {
		t.Fatal("expected AutoCreateIndex false")
	}
}

func TestLoadElasticsearchConfig_AutoCreateExplicit(t *testing.T) {
	t.Setenv("ELASTICSEARCH_AUTO_CREATE_INDEX", "true")
	c := LoadElasticsearchConfig("production")
	if !c.AutoCreateIndex {
		t.Fatal("expected AutoCreateIndex true when env set")
	}
}

func TestLoadElasticsearchConfig_Addresses(t *testing.T) {
	t.Setenv("ELASTICSEARCH_ENABLED", "true")
	t.Setenv("ELASTICSEARCH_URLS", " http://a:9200 , https://b:9200 ")
	c := LoadElasticsearchConfig("dev")
	if len(c.Addresses) != 2 || c.Addresses[0] != "http://a:9200" || c.Addresses[1] != "https://b:9200" {
		t.Fatalf("Addresses: %#v", c.Addresses)
	}
}
