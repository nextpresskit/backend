package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"

	"github.com/nextpresskit/backend/internal/config"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
)

// PostsIndex wraps Elasticsearch operations for the posts search index.
type PostsIndex struct {
	client *elasticsearch.Client
	name   string
	log    *zap.SugaredLogger
}

// NewPostsIndex returns nil if client is nil.
func NewPostsIndex(client *elasticsearch.Client, cfg config.ElasticsearchConfig, log *zap.SugaredLogger) *PostsIndex {
	if client == nil || !cfg.Enabled {
		return nil
	}
	prefix := strings.TrimSpace(cfg.IndexPrefix)
	if prefix == "" {
		prefix = config.DefaultElasticsearchIndexPrefix
	}
	return &PostsIndex{
		client: client,
		name:   prefix + "_posts",
		log:    log,
	}
}

// Name returns the concrete index name.
func (p *PostsIndex) Name() string { return p.name }

// Ready verifies Elasticsearch reachability for readiness probes.
func (p *PostsIndex) Ready(ctx context.Context) error {
	if p == nil || p.client == nil {
		return nil
	}
	return Ping(ctx, p.client)
}

const postsIndexMapping = `{
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "title": { "type": "text" },
      "slug": { "type": "keyword" },
      "excerpt": { "type": "text" },
      "content": { "type": "text" },
      "status": { "type": "keyword" },
      "published_at": { "type": "date" }
    }
  }
}`

// EnsureIndex creates the posts index when missing (used when AutoCreateIndex is true).
func (p *PostsIndex) EnsureIndex(ctx context.Context) error {
	if p == nil || p.client == nil {
		return nil
	}
	res, err := p.client.Indices.Exists([]string{p.name})
	if err != nil {
		return fmt.Errorf("elasticsearch exists check: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}
	if res.StatusCode != 404 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch exists unexpected status %d: %s", res.StatusCode, string(body))
	}

	cres, err := p.client.Indices.Create(
		p.name,
		p.client.Indices.Create.WithBody(strings.NewReader(postsIndexMapping)),
		p.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elasticsearch create index: %w", err)
	}
	defer cres.Body.Close()
	if cres.IsError() {
		b, _ := io.ReadAll(cres.Body)
		return fmt.Errorf("elasticsearch create index failed: %s", string(b))
	}
	return nil
}

type postDoc struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

func shouldIndexPost(p *model.Post) bool {
	if p == nil {
		return false
	}
	return p.Status == ident.StatusPublished && p.DeletedAt == nil
}

// SyncPost indexes the post when published, otherwise removes it from the index.
func (p *PostsIndex) SyncPost(ctx context.Context, post *model.Post) {
	if p == nil || p.client == nil {
		return
	}
	id := strings.TrimSpace(post.UUID)
	if id == "" {
		return
	}
	if !shouldIndexPost(post) {
		p.DeletePost(ctx, id)
		return
	}
	doc := postDoc{
		ID:          id,
		Title:       post.Title,
		Slug:        post.Slug,
		Excerpt:     post.Excerpt,
		Content:     post.Content,
		Status:      string(post.Status),
		PublishedAt: post.PublishedAt,
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		p.log.Errorw("elasticsearch index marshal failed", "post_id", id, "error", err)
		return
	}
	res, err := p.client.Index(
		p.name,
		bytes.NewReader(payload),
		p.client.Index.WithDocumentID(id),
		p.client.Index.WithContext(ctx),
		p.client.Index.WithRefresh("false"),
	)
	if err != nil {
		p.log.Errorw("elasticsearch index request failed", "post_id", id, "error", err)
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		p.log.Errorw("elasticsearch index failed", "post_id", id, "response", string(b))
	}
}

// DeletePost removes a document by post id.
func (p *PostsIndex) DeletePost(ctx context.Context, postID string) {
	if p == nil || p.client == nil {
		return
	}
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return
	}
	res, err := p.client.Delete(
		p.name,
		postID,
		p.client.Delete.WithContext(ctx),
		p.client.Delete.WithRefresh("false"),
	)
	if err != nil {
		p.log.Errorw("elasticsearch delete request failed", "post_id", postID, "error", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return
	}
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		p.log.Errorw("elasticsearch delete failed", "post_id", postID, "response", string(b))
	}
}

// SearchPostIDs runs a multi_match query and returns post ids in score order.
func (p *PostsIndex) SearchPostIDs(ctx context.Context, q string, limit, offset int) ([]string, error) {
	if p == nil || p.client == nil {
		return nil, fmt.Errorf("elasticsearch not configured")
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{
					map[string]any{"term": map[string]any{"status": string(ident.StatusPublished)}},
					map[string]any{
						"multi_match": map[string]any{
							"query":  q,
							"fields": []string{"title^2", "excerpt", "content"},
							"type":   "best_fields",
						},
					},
				},
			},
		},
		"sort": []any{
			map[string]any{"published_at": map[string]any{"order": "desc", "missing": "_last"}},
		},
		"_source": false,
		"from":    offset,
		"size":    limit,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Search(
		p.client.Search.WithContext(ctx),
		p.client.Search.WithIndex(p.name),
		p.client.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("elasticsearch search: %s", string(b))
	}
	var parsed struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(parsed.Hits.Hits))
	for _, h := range parsed.Hits.Hits {
		if h.ID != "" {
			out = append(out, h.ID)
		}
	}
	return out, nil
}
