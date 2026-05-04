package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type Handler struct {
	core      PostsCore
	sub       PostsSubresources
	series    SeriesAdmin
	groups    TranslationGroupsAdmin
	esBackend PostsElasticsearchBackend
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	posts := rg.Group("/posts")

	posts.GET("",
		auth,
		requirePerm("posts:read"),
		h.list,
	)
	posts.GET("/:id",
		auth,
		requirePerm("posts:read"),
		h.getByID,
	)
	posts.POST("",
		auth,
		requirePerm("posts:write"),
		h.create,
	)
	posts.POST("/search/reindex",
		auth,
		requirePerm("posts:write"),
		h.adminSearchReindex,
	)
	posts.PUT("/:id",
		auth,
		requirePerm("posts:write"),
		h.update,
	)
	posts.PUT("/:id/metrics",
		auth,
		requirePerm("posts:write"),
		h.setMetrics,
	)
	posts.DELETE("/:id",
		auth,
		requirePerm("posts:write"),
		h.delete,
	)

	posts.PUT("/:id/categories",
		auth,
		requirePerm("posts:write"),
		h.setCategories,
	)
	posts.PUT("/:id/tags",
		auth,
		requirePerm("posts:write"),
		h.setTags,
	)
	posts.PUT("/:id/primary-category",
		auth,
		requirePerm("posts:write"),
		h.setPrimaryCategory,
	)

	h.registerPostSubresourceRoutes(posts, auth, requirePerm)
	h.registerSeriesAndTranslationRoutes(rg, auth, requirePerm)
}

type setPrimaryCategoryRequest struct {
	CategoryID *string `json:"categoryId"`
}

func (h *Handler) setPrimaryCategory(c *gin.Context) {
	id := c.Param("id")
	var req setPrimaryCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.core.SetPrimaryCategory(c.Request.Context(), id, req.CategoryID); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, postToJSON(p))
}

type setMetricsRequest struct {
	Metrics struct {
		WordCount             *int     `json:"wordCount"`
		CharacterCount        *int     `json:"characterCount"`
		ReadingTimeMinutes    *int     `json:"readingTimeMinutes"`
		EstReadTimeSeconds    *int     `json:"estReadTimeSeconds"`
		ViewCount             *int64   `json:"viewCount"`
		UniqueVisitors7d      *int64   `json:"uniqueVisitors7d"`
		ScrollDepthAvgPercent *float32 `json:"scrollDepthAvgPercent"`
		BounceRatePercent     *float32 `json:"bounceRatePercent"`
		AvgTimeOnPageSeconds  *int     `json:"avgTimeOnPageSeconds"`
		CommentCount          *int     `json:"commentCount"`
		LikeCount             *int     `json:"likeCount"`
		ShareCount            *int     `json:"shareCount"`
		BookmarkCount         *int     `json:"bookmarkCount"`
	} `json:"metrics" binding:"required"`
}

func (h *Handler) setMetrics(c *gin.Context) {
	id := c.Param("id")

	var req setMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	if p.Metrics == nil {
		p.Metrics = &metrics.PostMetrics{}
	}
	if req.Metrics.WordCount != nil {
		p.Metrics.WordCount = *req.Metrics.WordCount
	}
	if req.Metrics.CharacterCount != nil {
		p.Metrics.CharacterCount = *req.Metrics.CharacterCount
	}
	if req.Metrics.ReadingTimeMinutes != nil {
		p.Metrics.ReadingTimeMinutes = *req.Metrics.ReadingTimeMinutes
	}
	if req.Metrics.EstReadTimeSeconds != nil {
		p.Metrics.EstReadTimeSeconds = *req.Metrics.EstReadTimeSeconds
	}
	if req.Metrics.ViewCount != nil {
		p.Metrics.ViewCount = *req.Metrics.ViewCount
	}
	if req.Metrics.UniqueVisitors7d != nil {
		p.Metrics.UniqueVisitors7d = *req.Metrics.UniqueVisitors7d
	}
	if req.Metrics.ScrollDepthAvgPercent != nil {
		p.Metrics.ScrollDepthAvgPercent = *req.Metrics.ScrollDepthAvgPercent
	}
	if req.Metrics.BounceRatePercent != nil {
		p.Metrics.BounceRatePercent = *req.Metrics.BounceRatePercent
	}
	if req.Metrics.AvgTimeOnPageSeconds != nil {
		p.Metrics.AvgTimeOnPageSeconds = *req.Metrics.AvgTimeOnPageSeconds
	}
	if req.Metrics.CommentCount != nil {
		p.Metrics.CommentCount = *req.Metrics.CommentCount
	}
	if req.Metrics.LikeCount != nil {
		p.Metrics.LikeCount = *req.Metrics.LikeCount
	}
	if req.Metrics.ShareCount != nil {
		p.Metrics.ShareCount = *req.Metrics.ShareCount
	}
	if req.Metrics.BookmarkCount != nil {
		p.Metrics.BookmarkCount = *req.Metrics.BookmarkCount
	}
	p.Metrics.UpdatedAt = time.Now().UTC()

	p, err = h.core.Save(c.Request.Context(), p)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, postToJSON(p))
}

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	posts := rg.Group("/posts")
	posts.GET("/search", h.publicSearch)
	posts.GET("", h.publicList)
	posts.GET("/:slug", h.publicGetBySlug)
}

type createPostRequest struct {
	Title   string `json:"title" binding:"required"`
	Slug    string `json:"slug" binding:"required"`
	Content string `json:"content"`

	UUID       *string `json:"uuid"`
	Subtitle   *string `json:"subtitle"`
	Excerpt    *string `json:"excerpt"`
	PostType   *string `json:"type"`
	Format     *string `json:"format"`
	Visibility *string `json:"visibility"`
	Locale     *string `json:"locale"`
	Timezone   *string `json:"timezone"`

	ReviewerUserID     *string `json:"reviewerUserId"`
	LastEditedByUserID *string `json:"lastEditedByUserId"`
	WorkflowStage      *string `json:"workflowStage"`
	Revision           *int    `json:"revision"`

	ScheduledPublishAt *string `json:"scheduledPublishAt"`
	FirstIndexedAt     *string `json:"firstIndexedAt"`

	CustomFields any `json:"customFields"`
	Flags        any `json:"flags"`
	Engagement   any `json:"engagement"`
	Workflow     any `json:"workflow"`

	FeaturedImage *struct {
		MediaID    *string `json:"mediaId"`
		URL        *string `json:"url"`
		Alt        *string `json:"alt"`
		Width      *int    `json:"width"`
		Height     *int    `json:"height"`
		FocalPoint *struct {
			X *float32 `json:"x"`
			Y *float32 `json:"y"`
		} `json:"focalPoint"`
		Credit  *string `json:"credit"`
		License *string `json:"license"`
	} `json:"featuredImage"`

	SEO *struct {
		Title          *string `json:"title"`
		Description    *string `json:"description"`
		CanonicalURL   *string `json:"canonicalUrl"`
		Robots         *string `json:"robots"`
		OGType         *string `json:"ogType"`
		OGImage        *string `json:"ogImage"`
		TwitterCard    *string `json:"twitterCard"`
		StructuredData any     `json:"structuredData"`
	} `json:"seo"`

	Metrics *struct {
		WordCount             *int     `json:"wordCount"`
		CharacterCount        *int     `json:"characterCount"`
		ReadingTimeMinutes    *int     `json:"readingTimeMinutes"`
		EstReadTimeSeconds    *int     `json:"estReadTimeSeconds"`
		ViewCount             *int64   `json:"viewCount"`
		UniqueVisitors7d      *int64   `json:"uniqueVisitors7d"`
		ScrollDepthAvgPercent *float32 `json:"scrollDepthAvgPercent"`
		BounceRatePercent     *float32 `json:"bounceRatePercent"`
		AvgTimeOnPageSeconds  *int     `json:"avgTimeOnPageSeconds"`
		CommentCount          *int     `json:"commentCount"`
		LikeCount             *int     `json:"likeCount"`
		ShareCount            *int     `json:"shareCount"`
		BookmarkCount         *int     `json:"bookmarkCount"`
	} `json:"metrics"`
}

func (h *Handler) create(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	v, _ := c.Get(platformMiddleware.ContextUserIDKey)
	authorID, _ := v.(string)

	p, err := h.core.Create(c.Request.Context(), authorID, req.Title, req.Slug, req.Content)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	applyCreateExtrasFromRequest(p, &req)
	p, err = h.core.Save(c.Request.Context(), p)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, postToJSON(p))
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, postToJSON(p))
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	status := c.Query("status")
	authorID := c.Query("author_id")
	q := c.Query("q")

	posts, err := h.core.ListFiltered(c.Request.Context(), limit, offset, status, authorID, q)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	out := make([]gin.H, 0, len(posts))
	for i := range posts {
		p := posts[i]
		out = append(out, postToJSON(&p))
	}
	c.JSON(http.StatusOK, gin.H{"posts": out})
}

func (h *Handler) publicList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	q := c.Query("q")
	categoryID := c.Query("category_id")
	tagID := c.Query("tag_id")

	posts, err := h.core.PublicList(c.Request.Context(), limit, offset, q, categoryID, tagID)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	out := make([]gin.H, 0, len(posts))
	for i := range posts {
		p := posts[i]
		out = append(out, postToJSON(&p))
	}
	c.JSON(http.StatusOK, gin.H{"posts": out})
}

func (h *Handler) adminSearchReindex(c *gin.Context) {
	if h.esBackend == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "search_disabled"})
		return
	}
	n, err := h.core.ReindexPublishedForSearch(c.Request.Context(), func(ctx context.Context, p *model.Post) {
		h.esBackend.SyncPost(ctx, p)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reindex_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reindexed": n})
}

func (h *Handler) publicSearch(c *gin.Context) {
	if h.esBackend == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "search_disabled"})
		return
	}
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_query"})
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	ids, err := h.esBackend.SearchPostIDs(c.Request.Context(), q, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search_failed"})
		return
	}
	out := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		p, err := h.core.GetByID(c.Request.Context(), id)
		if err != nil || p == nil {
			continue
		}
		if p.Status != ident.StatusPublished || p.DeletedAt != nil {
			continue
		}
		out = append(out, postToJSON(p))
	}
	c.JSON(http.StatusOK, gin.H{"posts": out})
}

func (h *Handler) publicGetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.core.PublicGetBySlug(c.Request.Context(), slug)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, postToJSON(p))
}

type updatePostRequest struct {
	Title   *string `json:"title"`
	Slug    *string `json:"slug"`
	Content *string `json:"content"`
	Status  *string `json:"status"`

	UUID       *string `json:"uuid"`
	Subtitle   *string `json:"subtitle"`
	Excerpt    *string `json:"excerpt"`
	PostType   *string `json:"type"`
	Format     *string `json:"format"`
	Visibility *string `json:"visibility"`
	Locale     *string `json:"locale"`
	Timezone   *string `json:"timezone"`

	ReviewerUserID     *string `json:"reviewerUserId"`
	LastEditedByUserID *string `json:"lastEditedByUserId"`
	WorkflowStage      *string `json:"workflowStage"`
	Revision           *int    `json:"revision"`

	ScheduledPublishAt *string `json:"scheduledPublishAt"`
	FirstIndexedAt     *string `json:"firstIndexedAt"`

	CustomFields any `json:"customFields"`
	Flags        any `json:"flags"`
	Engagement   any `json:"engagement"`
	Workflow     any `json:"workflow"`

	FeaturedImage *struct {
		MediaID    *string `json:"mediaId"`
		URL        *string `json:"url"`
		Alt        *string `json:"alt"`
		Width      *int    `json:"width"`
		Height     *int    `json:"height"`
		FocalPoint *struct {
			X *float32 `json:"x"`
			Y *float32 `json:"y"`
		} `json:"focalPoint"`
		Credit  *string `json:"credit"`
		License *string `json:"license"`
	} `json:"featuredImage"`

	SEO *struct {
		Title          *string `json:"title"`
		Description    *string `json:"description"`
		CanonicalURL   *string `json:"canonicalUrl"`
		Robots         *string `json:"robots"`
		OGType         *string `json:"ogType"`
		OGImage        *string `json:"ogImage"`
		TwitterCard    *string `json:"twitterCard"`
		StructuredData any     `json:"structuredData"`
	} `json:"seo"`

	Metrics *struct {
		WordCount             *int     `json:"wordCount"`
		CharacterCount        *int     `json:"characterCount"`
		ReadingTimeMinutes    *int     `json:"readingTimeMinutes"`
		EstReadTimeSeconds    *int     `json:"estReadTimeSeconds"`
		ViewCount             *int64   `json:"viewCount"`
		UniqueVisitors7d      *int64   `json:"uniqueVisitors7d"`
		ScrollDepthAvgPercent *float32 `json:"scrollDepthAvgPercent"`
		BounceRatePercent     *float32 `json:"bounceRatePercent"`
		AvgTimeOnPageSeconds  *int     `json:"avgTimeOnPageSeconds"`
		CommentCount          *int     `json:"commentCount"`
		LikeCount             *int     `json:"likeCount"`
		ShareCount            *int     `json:"shareCount"`
		BookmarkCount         *int     `json:"bookmarkCount"`
	} `json:"metrics"`
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")

	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	title := ""
	if req.Title != nil {
		title = *req.Title
	}
	slug := ""
	if req.Slug != nil {
		slug = *req.Slug
	}
	content := ""
	if req.Content != nil {
		content = *req.Content
	}
	status := ""
	if req.Status != nil {
		status = *req.Status
	}

	p, err := h.core.Update(c.Request.Context(), id, title, slug, content, status)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	applyUpdateExtrasFromRequest(p, &req)
	// persist extras
	p, err = h.core.Save(c.Request.Context(), p)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, postToJSON(p))
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.core.Delete(c.Request.Context(), id); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type setIDsRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

func wantsExpandedPost(c *gin.Context) bool {
	return c.Query("expand") == "post"
}

func wantsExpandedSeries(c *gin.Context) bool {
	return c.Query("expand") == "series"
}

func (h *Handler) setCategories(c *gin.Context) {
	id := c.Param("id")

	var req setIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.core.SetCategories(c.Request.Context(), id, req.IDs); err != nil {
		respondPostsServiceError(c, err)
		return
	}

	if wantsExpandedPost(c) {
		h.respondFullPost(c, id, http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) setTags(c *gin.Context) {
	id := c.Param("id")

	var req setIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	if err := h.core.SetTags(c.Request.Context(), id, req.IDs); err != nil {
		respondPostsServiceError(c, err)
		return
	}

	if wantsExpandedPost(c) {
		h.respondFullPost(c, id, http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func postToJSON(p *model.Post) gin.H {
	var publishedAt any
	if p.PublishedAt != nil {
		publishedAt = p.PublishedAt
	}

	var scheduledPublishAt any
	if p.ScheduledPublishAt != nil {
		scheduledPublishAt = p.ScheduledPublishAt
	}
	var firstIndexedAt any
	if p.FirstIndexedAt != nil {
		firstIndexedAt = p.FirstIndexedAt
	}

	var uuid any
	if p.UUID != nil {
		uuid = *p.UUID
	}

	var featuredImage any
	if p.FeaturedMediaID != nil || p.FeaturedMediaPublicURL != nil || p.FeaturedAlt != nil || p.FeaturedWidth != nil || p.FeaturedHeight != nil || p.FeaturedFocalX != nil || p.FeaturedFocalY != nil || p.FeaturedCredit != nil || p.FeaturedLicense != nil {
		fp := gin.H{}
		if p.FeaturedFocalX != nil || p.FeaturedFocalY != nil {
			fp = gin.H{"x": p.FeaturedFocalX, "y": p.FeaturedFocalY}
		}
		featuredImage = gin.H{
			"mediaId":    p.FeaturedMediaID,
			"url":        p.FeaturedMediaPublicURL,
			"alt":        p.FeaturedAlt,
			"width":      p.FeaturedWidth,
			"height":     p.FeaturedHeight,
			"focalPoint": fp,
			"credit":     p.FeaturedCredit,
			"license":    p.FeaturedLicense,
		}
	}

	var primaryCategory any
	for i := range p.Categories {
		if p.Categories[i].IsPrimary {
			c := p.Categories[i]
			primaryCategory = &c
			break
		}
	}

	var seo any
	if p.SEO != nil {
		seo = gin.H{
			"title":          p.SEO.Title,
			"description":    p.SEO.Description,
			"canonicalUrl":   p.SEO.CanonicalURL,
			"robots":         p.SEO.Robots,
			"ogType":         p.SEO.OGType,
			"ogImage":        p.SEO.OGImageURL,
			"twitterCard":    p.SEO.TwitterCard,
			"structuredData": jsonToAny(p.SEO.StructuredData),
		}
	}

	var metrics any
	if p.Metrics != nil {
		metrics = gin.H{
			"wordCount":             p.Metrics.WordCount,
			"characterCount":        p.Metrics.CharacterCount,
			"readingTimeMinutes":    p.Metrics.ReadingTimeMinutes,
			"estReadTimeSeconds":    p.Metrics.EstReadTimeSeconds,
			"viewCount":             p.Metrics.ViewCount,
			"uniqueVisitors7d":      p.Metrics.UniqueVisitors7d,
			"scrollDepthAvgPercent": p.Metrics.ScrollDepthAvgPercent,
			"bounceRatePercent":     p.Metrics.BounceRatePercent,
			"avgTimeOnPageSeconds":  p.Metrics.AvgTimeOnPageSeconds,
			"commentCount":          p.Metrics.CommentCount,
			"likeCount":             p.Metrics.LikeCount,
			"shareCount":            p.Metrics.ShareCount,
			"bookmarkCount":         p.Metrics.BookmarkCount,
		}
	}

	var author any
	if p.Author != nil {
		author = p.Author
	}
	var reviewer any
	if p.Reviewer != nil {
		reviewer = p.Reviewer
	}
	var lastEditedBy any
	if p.LastEditedBy != nil {
		lastEditedBy = p.LastEditedBy
	}

	return gin.H{
		"id":                 p.ID,
		"uuid":               uuid,
		"authorId":           p.AuthorID,
		"author":             author,
		"title":              p.Title,
		"slug":               p.Slug,
		"subtitle":           p.Subtitle,
		"excerpt":            p.Excerpt,
		"type":               p.PostType,
		"format":             p.Format,
		"visibility":         p.Visibility,
		"locale":             p.Locale,
		"timezone":           p.Timezone,
		"content":            p.Content,
		"status":             p.Status,
		"workflowStage":      p.WorkflowStage,
		"revision":           p.Revision,
		"reviewerUserId":     p.ReviewerUserID,
		"reviewer":           reviewer,
		"lastEditedByUserId": p.LastEditedByUserID,
		"lastEditedBy":       lastEditedBy,
		"editors": func() any {
			if len(p.Editors) > 0 {
				return p.Editors
			}
			if len(p.EditorUserIDs) == 0 {
				return []any{}
			}
			out := make([]any, 0, len(p.EditorUserIDs))
			for _, id := range p.EditorUserIDs {
				out = append(out, gin.H{"id": id})
			}
			return out
		}(),
		"scheduledPublishAt": scheduledPublishAt,
		"publishedAt":        publishedAt,
		"firstIndexedAt":     firstIndexedAt,
		"customFields":       jsonToAny(p.CustomFields),
		"flags":              jsonToAny(p.Flags),
		"engagement":         jsonToAny(p.Engagement),
		"workflow": func() any {
			base := jsonToAny(p.Workflow)
			m := gin.H{}
			if mm, ok := base.(map[string]any); ok {
				for k, v := range mm {
					m[k] = v
				}
			}
			if len(p.Changelog) > 0 {
				m["changelog"] = p.Changelog
			}
			if p.WorkflowStage != "" {
				m["stage"] = p.WorkflowStage
			}
			return m
		}(),
		"categories": func() any {
			if len(p.Categories) == 0 {
				return []any{}
			}
			return p.Categories
		}(),
		"category": primaryCategory,
		"tags": func() any {
			if len(p.Tags) == 0 {
				return []any{}
			}
			return p.Tags
		}(),
		"series": p.Series,
		"syndication": func() any {
			if len(p.Syndication) == 0 {
				return []any{}
			}
			return p.Syndication
		}(),
		"translations": func() any {
			if p.Translations == nil {
				return []any{}
			}
			return p.Translations.Translations
		}(),
		"translationGroupId": func() any {
			if p.Translations == nil || p.Translations.GroupID == nil {
				return nil
			}
			return *p.Translations.GroupID
		}(),
		"gallery": func() any {
			if len(p.Gallery) == 0 {
				return []any{}
			}
			return p.Gallery
		}(),
		"featuredImage": featuredImage,
		"seo":           seo,
		"metrics":       metrics,
		"createdAt":     p.CreatedAt,
		"updatedAt":     p.UpdatedAt,
	}
}

func jsonToAny(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}
	return v
}

func anyToJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	tt := t.UTC()
	return &tt
}

func applyCreateExtrasFromRequest(p *model.Post, req *createPostRequest) {
	if req.UUID != nil {
		p.UUID = req.UUID
	}
	if req.Subtitle != nil {
		p.Subtitle = *req.Subtitle
	}
	if req.Excerpt != nil {
		p.Excerpt = *req.Excerpt
	}
	if req.PostType != nil {
		p.PostType = *req.PostType
	}
	if req.Format != nil {
		p.Format = *req.Format
	}
	if req.Visibility != nil {
		p.Visibility = *req.Visibility
	}
	if req.Locale != nil {
		p.Locale = *req.Locale
	}
	if req.Timezone != nil {
		p.Timezone = *req.Timezone
	}
	if req.ReviewerUserID != nil {
		p.ReviewerUserID = req.ReviewerUserID
	}
	if req.LastEditedByUserID != nil {
		p.LastEditedByUserID = req.LastEditedByUserID
	}
	if req.WorkflowStage != nil {
		p.WorkflowStage = *req.WorkflowStage
	}
	if req.Revision != nil {
		p.Revision = *req.Revision
	}
	p.ScheduledPublishAt = parseTimePtr(req.ScheduledPublishAt)
	p.FirstIndexedAt = parseTimePtr(req.FirstIndexedAt)
	p.CustomFields = anyToJSON(req.CustomFields)
	p.Flags = anyToJSON(req.Flags)
	p.Engagement = anyToJSON(req.Engagement)
	p.Workflow = anyToJSON(req.Workflow)
	if req.FeaturedImage != nil {
		p.FeaturedMediaID = req.FeaturedImage.MediaID
		p.FeaturedAlt = req.FeaturedImage.Alt
		p.FeaturedWidth = req.FeaturedImage.Width
		p.FeaturedHeight = req.FeaturedImage.Height
		if req.FeaturedImage.FocalPoint != nil {
			p.FeaturedFocalX = req.FeaturedImage.FocalPoint.X
			p.FeaturedFocalY = req.FeaturedImage.FocalPoint.Y
		}
		p.FeaturedCredit = req.FeaturedImage.Credit
		p.FeaturedLicense = req.FeaturedImage.License
	}
	if req.SEO != nil {
		p.SEO = &seo.PostSEO{
			Title:          req.SEO.Title,
			Description:    req.SEO.Description,
			CanonicalURL:   req.SEO.CanonicalURL,
			Robots:         req.SEO.Robots,
			OGType:         req.SEO.OGType,
			OGImageURL:     req.SEO.OGImage,
			TwitterCard:    req.SEO.TwitterCard,
			StructuredData: anyToJSON(req.SEO.StructuredData),
			UpdatedAt:      time.Now().UTC(),
		}
	}
	if req.Metrics != nil {
		m := &metrics.PostMetrics{}
		if req.Metrics.WordCount != nil {
			m.WordCount = *req.Metrics.WordCount
		}
		if req.Metrics.CharacterCount != nil {
			m.CharacterCount = *req.Metrics.CharacterCount
		}
		if req.Metrics.ReadingTimeMinutes != nil {
			m.ReadingTimeMinutes = *req.Metrics.ReadingTimeMinutes
		}
		if req.Metrics.EstReadTimeSeconds != nil {
			m.EstReadTimeSeconds = *req.Metrics.EstReadTimeSeconds
		}
		if req.Metrics.ViewCount != nil {
			m.ViewCount = *req.Metrics.ViewCount
		}
		if req.Metrics.UniqueVisitors7d != nil {
			m.UniqueVisitors7d = *req.Metrics.UniqueVisitors7d
		}
		if req.Metrics.ScrollDepthAvgPercent != nil {
			m.ScrollDepthAvgPercent = *req.Metrics.ScrollDepthAvgPercent
		}
		if req.Metrics.BounceRatePercent != nil {
			m.BounceRatePercent = *req.Metrics.BounceRatePercent
		}
		if req.Metrics.AvgTimeOnPageSeconds != nil {
			m.AvgTimeOnPageSeconds = *req.Metrics.AvgTimeOnPageSeconds
		}
		if req.Metrics.CommentCount != nil {
			m.CommentCount = *req.Metrics.CommentCount
		}
		if req.Metrics.LikeCount != nil {
			m.LikeCount = *req.Metrics.LikeCount
		}
		if req.Metrics.ShareCount != nil {
			m.ShareCount = *req.Metrics.ShareCount
		}
		if req.Metrics.BookmarkCount != nil {
			m.BookmarkCount = *req.Metrics.BookmarkCount
		}
		m.UpdatedAt = time.Now().UTC()
		p.Metrics = m
	}
}

func applyUpdateExtrasFromRequest(p *model.Post, req *updatePostRequest) {
	if req.UUID != nil {
		p.UUID = req.UUID
	}
	if req.Subtitle != nil {
		p.Subtitle = *req.Subtitle
	}
	if req.Excerpt != nil {
		p.Excerpt = *req.Excerpt
	}
	if req.PostType != nil {
		p.PostType = *req.PostType
	}
	if req.Format != nil {
		p.Format = *req.Format
	}
	if req.Visibility != nil {
		p.Visibility = *req.Visibility
	}
	if req.Locale != nil {
		p.Locale = *req.Locale
	}
	if req.Timezone != nil {
		p.Timezone = *req.Timezone
	}
	if req.ReviewerUserID != nil {
		p.ReviewerUserID = req.ReviewerUserID
	}
	if req.LastEditedByUserID != nil {
		p.LastEditedByUserID = req.LastEditedByUserID
	}
	if req.WorkflowStage != nil {
		p.WorkflowStage = *req.WorkflowStage
	}
	if req.Revision != nil {
		p.Revision = *req.Revision
	}
	if req.ScheduledPublishAt != nil {
		p.ScheduledPublishAt = parseTimePtr(req.ScheduledPublishAt)
	}
	if req.FirstIndexedAt != nil {
		p.FirstIndexedAt = parseTimePtr(req.FirstIndexedAt)
	}
	if req.CustomFields != nil {
		p.CustomFields = anyToJSON(req.CustomFields)
	}
	if req.Flags != nil {
		p.Flags = anyToJSON(req.Flags)
	}
	if req.Engagement != nil {
		p.Engagement = anyToJSON(req.Engagement)
	}
	if req.Workflow != nil {
		p.Workflow = anyToJSON(req.Workflow)
	}
	if req.FeaturedImage != nil {
		p.FeaturedMediaID = req.FeaturedImage.MediaID
		p.FeaturedAlt = req.FeaturedImage.Alt
		p.FeaturedWidth = req.FeaturedImage.Width
		p.FeaturedHeight = req.FeaturedImage.Height
		if req.FeaturedImage.FocalPoint != nil {
			p.FeaturedFocalX = req.FeaturedImage.FocalPoint.X
			p.FeaturedFocalY = req.FeaturedImage.FocalPoint.Y
		}
		p.FeaturedCredit = req.FeaturedImage.Credit
		p.FeaturedLicense = req.FeaturedImage.License
	}
	if req.SEO != nil {
		if p.SEO == nil {
			p.SEO = &seo.PostSEO{}
		}
		p.SEO.Title = req.SEO.Title
		p.SEO.Description = req.SEO.Description
		p.SEO.CanonicalURL = req.SEO.CanonicalURL
		p.SEO.Robots = req.SEO.Robots
		p.SEO.OGType = req.SEO.OGType
		p.SEO.OGImageURL = req.SEO.OGImage
		p.SEO.TwitterCard = req.SEO.TwitterCard
		p.SEO.StructuredData = anyToJSON(req.SEO.StructuredData)
		p.SEO.UpdatedAt = time.Now().UTC()
	}
	if req.Metrics != nil {
		if p.Metrics == nil {
			p.Metrics = &metrics.PostMetrics{}
		}
		if req.Metrics.WordCount != nil {
			p.Metrics.WordCount = *req.Metrics.WordCount
		}
		if req.Metrics.CharacterCount != nil {
			p.Metrics.CharacterCount = *req.Metrics.CharacterCount
		}
		if req.Metrics.ReadingTimeMinutes != nil {
			p.Metrics.ReadingTimeMinutes = *req.Metrics.ReadingTimeMinutes
		}
		if req.Metrics.EstReadTimeSeconds != nil {
			p.Metrics.EstReadTimeSeconds = *req.Metrics.EstReadTimeSeconds
		}
		if req.Metrics.ViewCount != nil {
			p.Metrics.ViewCount = *req.Metrics.ViewCount
		}
		if req.Metrics.UniqueVisitors7d != nil {
			p.Metrics.UniqueVisitors7d = *req.Metrics.UniqueVisitors7d
		}
		if req.Metrics.ScrollDepthAvgPercent != nil {
			p.Metrics.ScrollDepthAvgPercent = *req.Metrics.ScrollDepthAvgPercent
		}
		if req.Metrics.BounceRatePercent != nil {
			p.Metrics.BounceRatePercent = *req.Metrics.BounceRatePercent
		}
		if req.Metrics.AvgTimeOnPageSeconds != nil {
			p.Metrics.AvgTimeOnPageSeconds = *req.Metrics.AvgTimeOnPageSeconds
		}
		if req.Metrics.CommentCount != nil {
			p.Metrics.CommentCount = *req.Metrics.CommentCount
		}
		if req.Metrics.LikeCount != nil {
			p.Metrics.LikeCount = *req.Metrics.LikeCount
		}
		if req.Metrics.ShareCount != nil {
			p.Metrics.ShareCount = *req.Metrics.ShareCount
		}
		if req.Metrics.BookmarkCount != nil {
			p.Metrics.BookmarkCount = *req.Metrics.BookmarkCount
		}
		p.Metrics.UpdatedAt = time.Now().UTC()
	}
}
