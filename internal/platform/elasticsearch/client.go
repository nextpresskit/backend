package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"

	"github.com/nextpresskit/backend/internal/config"
)

// Ping checks cluster reachability (non-fatal at startup if it fails).
func Ping(ctx context.Context, client *elasticsearch.Client) error {
	if client == nil {
		return nil
	}
	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch: %s", string(b))
	}
	return nil
}

// NewClient builds an Elasticsearch client from config. Returns (nil, nil) when disabled.
func NewClient(cfg config.ElasticsearchConfig) (*elasticsearch.Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if len(cfg.Addresses) == 0 {
		return nil, nil
	}
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}
	if cfg.APIKey != "" {
		esCfg.APIKey = cfg.APIKey
	} else if cfg.Username != "" || cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}
	if cfg.InsecureSkipVerify {
		esCfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return elasticsearch.NewClient(esCfg)
}
