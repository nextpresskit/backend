package transport

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

func (h *Handler) registerPostSubresourceRoutes(posts *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	posts.GET("/:id/metrics",
		auth,
		requirePerm("posts:read"),
		h.adminGetMetrics,
	)
	posts.GET("/:id/seo",
		auth,
		requirePerm("posts:read"),
		h.adminGetSEO,
	)
	posts.PUT("/:id/seo",
		auth,
		requirePerm("posts:write"),
		h.adminPutSEO,
	)
	posts.DELETE("/:id/seo",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteSEO,
	)
	posts.PUT("/:id/featured-image",
		auth,
		requirePerm("posts:write"),
		h.adminPutFeaturedImage,
	)
	posts.PUT("/:id/series-link",
		auth,
		requirePerm("posts:write"),
		h.adminPutSeriesLink,
	)
	posts.GET("/:id/coauthors",
		auth,
		requirePerm("posts:read"),
		h.adminGetCoauthors,
	)
	posts.PUT("/:id/coauthors",
		auth,
		requirePerm("posts:write"),
		h.adminPutCoauthors,
	)
	posts.POST("/:id/gallery",
		auth,
		requirePerm("posts:write"),
		h.adminPostGalleryItem,
	)
	posts.PUT("/:id/gallery/:itemId",
		auth,
		requirePerm("posts:write"),
		h.adminPutGalleryItem,
	)
	posts.DELETE("/:id/gallery/:itemId",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteGalleryItem,
	)
	posts.POST("/:id/changelog",
		auth,
		requirePerm("posts:write"),
		h.adminPostChangelog,
	)
	posts.DELETE("/:id/changelog/:changelogId",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteChangelog,
	)
	posts.POST("/:id/syndication",
		auth,
		requirePerm("posts:write"),
		h.adminPostSyndication,
	)
	posts.PUT("/:id/syndication/:syndicationId",
		auth,
		requirePerm("posts:write"),
		h.adminPutSyndication,
	)
	posts.DELETE("/:id/syndication/:syndicationId",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteSyndication,
	)
	posts.PUT("/:id/translations",
		auth,
		requirePerm("posts:write"),
		h.adminPutTranslations,
	)
	posts.DELETE("/:id/translations",
		auth,
		requirePerm("posts:write"),
		h.adminClearTranslations,
	)
}

func (h *Handler) registerSeriesAndTranslationRoutes(rg *gin.RouterGroup, auth gin.HandlerFunc, requirePerm func(string) gin.HandlerFunc) {
	synd := rg.Group("/syndication")
	synd.PUT("/:id",
		auth,
		requirePerm("posts:write"),
		h.adminPutSyndicationGlobal,
	)
	synd.DELETE("/:id",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteSyndicationGlobal,
	)

	seriesG := rg.Group("/series")
	seriesG.GET("",
		auth,
		requirePerm("posts:read"),
		h.adminListSeries,
	)
	seriesG.POST("",
		auth,
		requirePerm("posts:write"),
		h.adminCreateSeries,
	)
	seriesG.GET("/:id",
		auth,
		requirePerm("posts:read"),
		h.adminGetSeries,
	)
	seriesG.PUT("/:id",
		auth,
		requirePerm("posts:write"),
		h.adminUpdateSeries,
	)
	seriesG.DELETE("/:id",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteSeries,
	)

	tg := rg.Group("/translation-groups")
	tg.POST("",
		auth,
		requirePerm("posts:write"),
		h.adminCreateTranslationGroup,
	)
	tg.GET("/:id",
		auth,
		requirePerm("posts:read"),
		h.adminGetTranslationGroup,
	)
	tg.DELETE("/:id",
		auth,
		requirePerm("posts:write"),
		h.adminDeleteTranslationGroup,
	)
}

func (h *Handler) respondFullPost(c *gin.Context, postID string, status int) {
	p, err := h.core.GetByID(c.Request.Context(), postID)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(status, postToJSON(p))
}

func metricsToGinH(m *metrics.PostMetrics) gin.H {
	if m == nil {
		return nil
	}
	return gin.H{
		"wordCount":             m.WordCount,
		"characterCount":        m.CharacterCount,
		"readingTimeMinutes":    m.ReadingTimeMinutes,
		"estReadTimeSeconds":    m.EstReadTimeSeconds,
		"viewCount":             m.ViewCount,
		"uniqueVisitors7d":      m.UniqueVisitors7d,
		"scrollDepthAvgPercent": m.ScrollDepthAvgPercent,
		"bounceRatePercent":     m.BounceRatePercent,
		"avgTimeOnPageSeconds":  m.AvgTimeOnPageSeconds,
		"commentCount":          m.CommentCount,
		"likeCount":             m.LikeCount,
		"shareCount":            m.ShareCount,
		"bookmarkCount":         m.BookmarkCount,
		"updatedAt":             m.UpdatedAt,
	}
}

func seoToGinH(doc *seo.PostSEO) any {
	if doc == nil {
		return nil
	}
	return gin.H{
		"title":          doc.Title,
		"description":    doc.Description,
		"canonicalUrl":   doc.CanonicalURL,
		"robots":         doc.Robots,
		"ogType":         doc.OGType,
		"ogImage":        doc.OGImageURL,
		"twitterCard":    doc.TwitterCard,
		"structuredData": jsonToAny(doc.StructuredData),
		"updatedAt":      doc.UpdatedAt,
	}
}

func seriesToGinH(s *series.Series) gin.H {
	if s == nil {
		return nil
	}
	return gin.H{
		"id":        s.ID,
		"title":     s.Title,
		"slug":      s.Slug,
		"createdAt": s.CreatedAt,
		"updatedAt": s.UpdatedAt,
	}
}

func (h *Handler) adminGetMetrics(c *gin.Context) {
	id := c.Param("id")
	m, err := h.sub.GetMetricsForPost(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"metrics": metricsToGinH(m)})
}

func (h *Handler) adminGetSEO(c *gin.Context) {
	id := c.Param("id")
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"seo": seoToGinH(p.SEO)})
}

type putSEORequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	CanonicalURL   *string `json:"canonicalUrl"`
	Robots         *string `json:"robots"`
	OGType         *string `json:"ogType"`
	OGImage        *string `json:"ogImage"`
	TwitterCard    *string `json:"twitterCard"`
	StructuredData any     `json:"structuredData"`
}

func (h *Handler) adminPutSEO(c *gin.Context) {
	id := c.Param("id")
	var req putSEORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	payload := &seo.PostSEO{
		Title:        req.Title,
		Description:  req.Description,
		CanonicalURL: req.CanonicalURL,
		Robots:       req.Robots,
		OGType:       req.OGType,
		OGImageURL:   req.OGImage,
		TwitterCard:  req.TwitterCard,
	}
	if req.StructuredData != nil {
		payload.StructuredData = anyToJSON(req.StructuredData)
	}
	if err := h.sub.UpsertSEO(c.Request.Context(), id, payload); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

func (h *Handler) adminDeleteSEO(c *gin.Context) {
	id := c.Param("id")
	if err := h.sub.DeleteSEO(c.Request.Context(), id); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

type putFeaturedImageRequest struct {
	MediaID    *string `json:"mediaId"`
	Alt        *string `json:"alt"`
	Width      *int    `json:"width"`
	Height     *int    `json:"height"`
	FocalPoint *struct {
		X *float32 `json:"x"`
		Y *float32 `json:"y"`
	} `json:"focalPoint"`
	Credit  *string `json:"credit"`
	License *string `json:"license"`
}

func (h *Handler) adminPutFeaturedImage(c *gin.Context) {
	id := c.Param("id")
	var req putFeaturedImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	var focalX, focalY *float32
	if req.FocalPoint != nil {
		focalX, focalY = req.FocalPoint.X, req.FocalPoint.Y
	}
	if err := h.sub.SetFeaturedImage(c.Request.Context(), id, req.MediaID, req.Alt, req.Width, req.Height, focalX, focalY, req.Credit, req.License); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

type putSeriesLinkRequest struct {
	SeriesID  *string `json:"seriesId"`
	PartIndex *int    `json:"partIndex"`
	PartLabel *string `json:"partLabel"`
}

func (h *Handler) adminPutSeriesLink(c *gin.Context) {
	id := c.Param("id")
	var req putSeriesLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.sub.SetPostSeries(c.Request.Context(), id, req.SeriesID, req.PartIndex, req.PartLabel); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

func (h *Handler) adminGetCoauthors(c *gin.Context) {
	id := c.Param("id")
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	out := make([]any, 0, len(p.CoAuthors))
	for i := range p.CoAuthors {
		out = append(out, p.CoAuthors[i])
	}
	c.JSON(http.StatusOK, gin.H{"coAuthors": out})
}

type putCoauthorsRequest struct {
	UserIDs []string `json:"userIds" binding:"required"`
}

func (h *Handler) adminPutCoauthors(c *gin.Context) {
	id := c.Param("id")
	var req putCoauthorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.sub.ReplaceCoauthors(c.Request.Context(), id, req.UserIDs); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

type postGalleryItemRequest struct {
	MediaID   string  `json:"mediaId" binding:"required"`
	SortOrder int     `json:"sortOrder"`
	Caption   *string `json:"caption"`
	Alt       *string `json:"alt"`
}

func (h *Handler) adminPostGalleryItem(c *gin.Context) {
	id := c.Param("id")
	var req postGalleryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	itemID, err := h.sub.CreateGalleryItem(c.Request.Context(), id, req.MediaID, req.SortOrder, req.Caption, req.Alt)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"itemId": itemID, "post": postToJSON(p)})
}

type putGalleryItemRequest struct {
	SortOrder *int    `json:"sortOrder"`
	Caption   *string `json:"caption"`
	Alt       *string `json:"alt"`
}

func (h *Handler) adminPutGalleryItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")
	var req putGalleryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.sub.UpdateGalleryItem(c.Request.Context(), id, itemID, req.SortOrder, req.Caption, req.Alt); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

func (h *Handler) adminDeleteGalleryItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")
	if err := h.sub.DeleteGalleryItem(c.Request.Context(), id, itemID); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

type postChangelogRequest struct {
	UserID *string `json:"userId"`
	Note   string  `json:"note" binding:"required"`
}

func (h *Handler) adminPostChangelog(c *gin.Context) {
	id := c.Param("id")
	var req postChangelogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	uid := req.UserID
	if uid == nil {
		v, _ := c.Get(platformMiddleware.ContextUserIDKey)
		if s, ok := v.(string); ok && s != "" {
			uid = &s
		}
	}
	changelogID, err := h.sub.CreateChangelog(c.Request.Context(), id, uid, req.Note)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"changelogId": changelogID, "post": postToJSON(p)})
}

func (h *Handler) adminDeleteChangelog(c *gin.Context) {
	id := c.Param("id")
	changelogID := c.Param("changelogId")
	if err := h.sub.DeleteChangelog(c.Request.Context(), id, changelogID); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

type postSyndicationRequest struct {
	Platform string `json:"platform" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Status   string `json:"status"`
}

func (h *Handler) adminPostSyndication(c *gin.Context) {
	id := c.Param("id")
	var req postSyndicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	syndID, err := h.sub.CreateSyndication(c.Request.Context(), id, req.Platform, req.URL, req.Status)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"syndicationId": syndID, "post": postToJSON(p)})
}

type putSyndicationRequest struct {
	Platform *string `json:"platform"`
	URL      *string `json:"url"`
	Status   *string `json:"status"`
}

func (h *Handler) adminPutSyndication(c *gin.Context) {
	id := c.Param("id")
	syndicationID := c.Param("syndicationId")
	var req putSyndicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.sub.UpdateSyndication(c.Request.Context(), id, syndicationID, req.Platform, req.URL, req.Status); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

func (h *Handler) adminDeleteSyndication(c *gin.Context) {
	id := c.Param("id")
	syndicationID := c.Param("syndicationId")
	if err := h.sub.DeleteSyndication(c.Request.Context(), id, syndicationID); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

func (h *Handler) adminPutSyndicationGlobal(c *gin.Context) {
	syndicationID := c.Param("id")
	var req putSyndicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	if err := h.sub.UpdateSyndicationGlobal(c.Request.Context(), syndicationID, req.Platform, req.URL, req.Status); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) adminDeleteSyndicationGlobal(c *gin.Context) {
	syndicationID := c.Param("id")
	if err := h.sub.DeleteSyndicationGlobal(c.Request.Context(), syndicationID); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type putTranslationsRequest struct {
	GroupID *string `json:"groupId"`
	Locale  string  `json:"locale" binding:"required"`
}

func (h *Handler) adminPutTranslations(c *gin.Context) {
	id := c.Param("id")
	var req putTranslationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	gid, err := h.sub.PutPostTranslation(c.Request.Context(), id, req.GroupID, req.Locale)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	p, err := h.core.GetByID(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"translationGroupId": gid, "post": postToJSON(p)})
}

func (h *Handler) adminClearTranslations(c *gin.Context) {
	id := c.Param("id")
	if err := h.sub.ClearPostTranslation(c.Request.Context(), id); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	h.respondFullPost(c, id, http.StatusOK)
}

func (h *Handler) adminListSeries(c *gin.Context) {
	list, err := h.series.ListSeries(c.Request.Context())
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for i := range list {
		out = append(out, seriesToGinH(&list[i]))
	}
	c.JSON(http.StatusOK, gin.H{"series": out})
}

type createSeriesRequest struct {
	Title string `json:"title" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
}

func (h *Handler) adminCreateSeries(c *gin.Context) {
	var req createSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	sr, err := h.series.CreateSeries(c.Request.Context(), req.Title, req.Slug)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, seriesToGinH(sr))
}

func (h *Handler) adminGetSeries(c *gin.Context) {
	id := c.Param("id")
	sr, err := h.series.GetSeries(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, seriesToGinH(sr))
}

type updateSeriesRequest struct {
	Title *string `json:"title"`
	Slug  *string `json:"slug"`
}

func (h *Handler) adminUpdateSeries(c *gin.Context) {
	id := c.Param("id")
	var req updateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}
	sr, err := h.series.UpdateSeries(c.Request.Context(), id, req.Title, req.Slug)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, seriesToGinH(sr))
}

func (h *Handler) adminDeleteSeries(c *gin.Context) {
	id := c.Param("id")
	var deletedSeries gin.H
	if wantsExpandedSeries(c) {
		sr, err := h.series.GetSeries(c.Request.Context(), id)
		if err != nil {
			respondPostsServiceError(c, err)
			return
		}
		deletedSeries = seriesToGinH(sr)
	}
	if err := h.series.DeleteSeries(c.Request.Context(), id); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	if deletedSeries != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true, "series": deletedSeries})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type createTranslationGroupRequest struct {
	ID *string `json:"id"`
}

func (h *Handler) adminCreateTranslationGroup(c *gin.Context) {
	var req createTranslationGroupRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
			return
		}
	}
	gid, err := h.groups.CreateTranslationGroup(c.Request.Context(), req.ID)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": gid})
}

func (h *Handler) adminGetTranslationGroup(c *gin.Context) {
	id := c.Param("id")
	ok, err := h.groups.TranslationGroupExists(c.Request.Context(), id)
	if err != nil {
		respondPostsServiceError(c, err)
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) adminDeleteTranslationGroup(c *gin.Context) {
	id := c.Param("id")
	if err := h.groups.DeleteTranslationGroup(c.Request.Context(), id); err != nil {
		respondPostsServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
