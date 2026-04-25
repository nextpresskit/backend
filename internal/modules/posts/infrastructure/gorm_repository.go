package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/ident"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/metrics"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/model"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/ports"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/seo"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/series"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormUserSummary struct {
	ID        int64  `gorm:"column:public_id"`
	UUID      string `gorm:"column:uuid"`
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	Email     string `gorm:"column:email"`
}

type gormPost struct {
	ID                 string          `gorm:"column:id;type:uuid;primaryKey"`
	UUID               *string         `gorm:"column:uuid;type:uuid;uniqueIndex"`
	AuthorID           int64           `gorm:"column:author_id;not null;index"`
	Title              string          `gorm:"column:title;not null"`
	Slug               string          `gorm:"column:slug;not null;uniqueIndex"`
	Subtitle           string          `gorm:"column:subtitle"`
	Excerpt            string          `gorm:"column:excerpt"`
	PostType           string          `gorm:"column:post_type"`
	Format             string          `gorm:"column:format"`
	Visibility         string          `gorm:"column:visibility;not null"`
	Locale             string          `gorm:"column:locale;not null"`
	Timezone           string          `gorm:"column:timezone;not null"`
	Content            string          `gorm:"column:content;not null"`
	Status             string          `gorm:"column:status;not null"`
	WorkflowStage      string          `gorm:"column:workflow_stage;not null"`
	Revision           int             `gorm:"column:revision;not null"`
	ReviewerUserID     *int64          `gorm:"column:reviewer_user_id"`
	LastEditedByUserID *int64          `gorm:"column:last_edited_by_user_id"`
	ScheduledPublishAt *time.Time      `gorm:"column:scheduled_publish_at"`
	PublishedAt        *time.Time      `gorm:"column:published_at"`
	FirstIndexedAt     *time.Time      `gorm:"column:first_indexed_at"`
	CustomFields       json.RawMessage `gorm:"column:custom_fields;type:jsonb;not null"`
	Flags              json.RawMessage `gorm:"column:flags;type:jsonb;not null"`
	Engagement         json.RawMessage `gorm:"column:engagement;type:jsonb;not null"`
	Workflow           json.RawMessage `gorm:"column:workflow;type:jsonb;not null"`
	FeaturedMediaID    *string         `gorm:"column:featured_media_id;type:uuid"`
	FeaturedAlt        *string         `gorm:"column:featured_alt"`
	FeaturedWidth      *int            `gorm:"column:featured_width"`
	FeaturedHeight     *int            `gorm:"column:featured_height"`
	FeaturedFocalX     *float32        `gorm:"column:featured_focal_x"`
	FeaturedFocalY     *float32        `gorm:"column:featured_focal_y"`
	FeaturedCredit     *string         `gorm:"column:featured_credit"`
	FeaturedLicense    *string         `gorm:"column:featured_license"`
	PrimaryCategoryID  *string         `gorm:"column:primary_category_id;type:uuid"`
	CreatedAt          time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time       `gorm:"column:updated_at;not null"`
	DeletedAt          gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

func (gormPost) TableName() string { return "posts" }

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, post *model.Post) error {
	m := fromDomain(post)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		if err := upsertSEO(ctx, tx, ident.PostID(m.ID), post.SEO); err != nil {
			return err
		}
		if err := upsertMetrics(ctx, tx, ident.PostID(m.ID), post.Metrics); err != nil {
			return err
		}
		loaded, err := r.findByIDWithExtras(ctx, tx, ident.PostID(m.ID))
		if err != nil {
			return err
		}
		*post = *loaded
		return nil
	})
}

func (r *GormRepository) FindByID(ctx context.Context, id ident.PostID) (*model.Post, error) {
	return r.findByIDWithExtras(ctx, r.db.WithContext(ctx), id)
}

func (r *GormRepository) FindBySlug(ctx context.Context, slug string) (*model.Post, error) {
	var row gormPost
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.findByIDWithExtras(ctx, r.db.WithContext(ctx), ident.PostID(row.ID))
}

func (r *GormRepository) List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]model.Post, error) {
	return r.ListFiltered(ctx, includeDeleted, limit, offset, "", "", "")
}

func (r *GormRepository) ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]model.Post, error) {
	dbq := r.db.WithContext(ctx).Model(&gormPost{}).Order("created_at DESC").Limit(limit).Offset(offset)
	if includeDeleted {
		dbq = dbq.Unscoped()
	}

	if status != "" {
		dbq = dbq.Where("status = ?", status)
	}
	if authorID != "" {
		dbq = dbq.Where("author_id = ?", authorID)
	}
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("(title ILIKE ? OR content ILIKE ?)", like, like)
	}

	var rows []gormPost
	if err := dbq.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]model.Post, 0, len(rows))
	for i := range rows {
		p := toDomain(&rows[i])
		out = append(out, *p)
	}
	return out, nil
}

func (r *GormRepository) ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]model.Post, error) {
	dbq := r.db.WithContext(ctx).
		Model(&gormPost{}).
		Where("status = ?", string(ident.StatusPublished)).
		Order("published_at DESC").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("(title ILIKE ? OR content ILIKE ?)", like, like)
	}
	if categoryID != "" {
		dbq = dbq.Joins("JOIN post_categories pc ON pc.post_id = posts.id").
			Where("pc.category_id = ?", categoryID)
	}
	if tagID != "" {
		dbq = dbq.Joins("JOIN post_tags pt ON pt.post_id = posts.id").
			Where("pt.tag_id = ?", tagID)
	}

	var rows []gormPost
	if err := dbq.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]model.Post, 0, len(rows))
	for i := range rows {
		p := toDomain(&rows[i])
		out = append(out, *p)
	}
	return out, nil
}

func (r *GormRepository) FindPublishedBySlug(ctx context.Context, slug string) (*model.Post, error) {
	var row gormPost
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		Where("status = ?", string(ident.StatusPublished)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.findByIDWithExtras(ctx, r.db.WithContext(ctx), ident.PostID(row.ID))
}

func (r *GormRepository) Update(ctx context.Context, post *model.Post) error {
	m := fromDomain(post)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&gormPost{}).
			Where("id = ?", m.ID).
			Updates(m).Error; err != nil {
			return err
		}
		if err := upsertSEO(ctx, tx, ident.PostID(m.ID), post.SEO); err != nil {
			return err
		}
		if err := upsertMetrics(ctx, tx, ident.PostID(m.ID), post.Metrics); err != nil {
			return err
		}
		return nil
	})
}

func (r *GormRepository) Delete(ctx context.Context, id ident.PostID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", string(id)).
		Delete(&gormPost{}).Error
}

type gormPostCategory struct {
	PostID     string `gorm:"column:post_id;type:uuid;primaryKey"`
	CategoryID string `gorm:"column:category_id;type:uuid;primaryKey"`
}

func (gormPostCategory) TableName() string { return "post_categories" }

type gormPostTag struct {
	PostID string `gorm:"column:post_id;type:uuid;primaryKey"`
	TagID  string `gorm:"column:tag_id;type:uuid;primaryKey"`
}

func (gormPostTag) TableName() string { return "post_tags" }

func (r *GormRepository) SetCategories(ctx context.Context, postID ident.PostID, categoryIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostCategory{}).Error; err != nil {
			return err
		}
		if len(categoryIDs) == 0 {
			return nil
		}

		rows := make([]gormPostCategory, 0, len(categoryIDs))
		for _, id := range categoryIDs {
			if id == "" {
				continue
			}
			rows = append(rows, gormPostCategory{PostID: string(postID), CategoryID: id})
		}
		if len(rows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func (r *GormRepository) SetTags(ctx context.Context, postID ident.PostID, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostTag{}).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}

		rows := make([]gormPostTag, 0, len(tagIDs))
		for _, id := range tagIDs {
			if id == "" {
				continue
			}
			rows = append(rows, gormPostTag{PostID: string(postID), TagID: id})
		}
		if len(rows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func (r *GormRepository) SetPrimaryCategory(ctx context.Context, postID ident.PostID, categoryID *string) error {
	return r.db.WithContext(ctx).
		Model(&gormPost{}).
		Where("id = ?", string(postID)).
		Update("primary_category_id", categoryID).Error
}

func toDomain(m *gormPost) *model.Post {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &model.Post{
		ID:                 ident.PostID(m.ID),
		UUID:               m.UUID,
		AuthorID:           strconv.FormatInt(m.AuthorID, 10),
		Title:              m.Title,
		Slug:               m.Slug,
		Subtitle:           m.Subtitle,
		Excerpt:            m.Excerpt,
		PostType:           m.PostType,
		Format:             m.Format,
		Visibility:         m.Visibility,
		Locale:             m.Locale,
		Timezone:           m.Timezone,
		Content:            m.Content,
		Status:             ident.Status(m.Status),
		WorkflowStage:      m.WorkflowStage,
		Revision:           m.Revision,
		ReviewerUserID:     int64PtrToStringPtr(m.ReviewerUserID),
		LastEditedByUserID: int64PtrToStringPtr(m.LastEditedByUserID),
		ScheduledPublishAt: m.ScheduledPublishAt,
		PublishedAt:        m.PublishedAt,
		FirstIndexedAt:     m.FirstIndexedAt,
		CustomFields:       m.CustomFields,
		Flags:              m.Flags,
		Engagement:         m.Engagement,
		Workflow:           m.Workflow,
		FeaturedMediaID:    m.FeaturedMediaID,
		FeaturedAlt:        m.FeaturedAlt,
		FeaturedWidth:      m.FeaturedWidth,
		FeaturedHeight:     m.FeaturedHeight,
		FeaturedFocalX:     m.FeaturedFocalX,
		FeaturedFocalY:     m.FeaturedFocalY,
		FeaturedCredit:     m.FeaturedCredit,
		FeaturedLicense:    m.FeaturedLicense,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		DeletedAt:          deletedAt,
	}
}

func fromDomain(p *model.Post) *gormPost {
	var deleted gorm.DeletedAt
	if p.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *p.DeletedAt, Valid: true}
	}
	return &gormPost{
		ID:                 string(p.ID),
		UUID:               p.UUID,
		AuthorID:           parseIDString(p.AuthorID),
		Title:              p.Title,
		Slug:               p.Slug,
		Subtitle:           p.Subtitle,
		Excerpt:            p.Excerpt,
		PostType:           p.PostType,
		Format:             p.Format,
		Visibility:         p.Visibility,
		Locale:             p.Locale,
		Timezone:           p.Timezone,
		Content:            p.Content,
		Status:             string(p.Status),
		WorkflowStage:      p.WorkflowStage,
		Revision:           p.Revision,
		ReviewerUserID:     parseOptionalIDString(p.ReviewerUserID),
		LastEditedByUserID: parseOptionalIDString(p.LastEditedByUserID),
		ScheduledPublishAt: p.ScheduledPublishAt,
		PublishedAt:        p.PublishedAt,
		FirstIndexedAt:     p.FirstIndexedAt,
		CustomFields:       nonNilJSON(p.CustomFields),
		Flags:              nonNilJSON(p.Flags),
		Engagement:         nonNilJSON(p.Engagement),
		Workflow:           nonNilJSON(p.Workflow),
		FeaturedMediaID:    p.FeaturedMediaID,
		FeaturedAlt:        p.FeaturedAlt,
		FeaturedWidth:      p.FeaturedWidth,
		FeaturedHeight:     p.FeaturedHeight,
		FeaturedFocalX:     p.FeaturedFocalX,
		FeaturedFocalY:     p.FeaturedFocalY,
		FeaturedCredit:     p.FeaturedCredit,
		FeaturedLicense:    p.FeaturedLicense,
		PrimaryCategoryID:  nil,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		DeletedAt:          deleted,
	}
}

type gormPostSEO struct {
	PostID         string          `gorm:"column:post_id;type:uuid;primaryKey"`
	Title          *string         `gorm:"column:title"`
	Description    *string         `gorm:"column:description"`
	CanonicalURL   *string         `gorm:"column:canonical_url"`
	Robots         *string         `gorm:"column:robots"`
	OGType         *string         `gorm:"column:og_type"`
	OGImageURL     *string         `gorm:"column:og_image_url"`
	TwitterCard    *string         `gorm:"column:twitter_card"`
	StructuredData json.RawMessage `gorm:"column:structured_data;type:jsonb;not null"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;not null"`
}

func (gormPostSEO) TableName() string { return "post_seo" }

type gormPostMetrics struct {
	PostID                string    `gorm:"column:post_id;type:uuid;primaryKey"`
	WordCount             int       `gorm:"column:word_count;not null"`
	CharacterCount        int       `gorm:"column:character_count;not null"`
	ReadingTimeMinutes    int       `gorm:"column:reading_time_minutes;not null"`
	EstReadTimeSeconds    int       `gorm:"column:est_read_time_seconds;not null"`
	ViewCount             int64     `gorm:"column:view_count;not null"`
	UniqueVisitors7d      int64     `gorm:"column:unique_visitors_7d;not null"`
	ScrollDepthAvgPercent float32   `gorm:"column:scroll_depth_avg_percent;not null"`
	BounceRatePercent     float32   `gorm:"column:bounce_rate_percent;not null"`
	AvgTimeOnPageSeconds  int       `gorm:"column:avg_time_on_page_seconds;not null"`
	CommentCount          int       `gorm:"column:comment_count;not null"`
	LikeCount             int       `gorm:"column:like_count;not null"`
	ShareCount            int       `gorm:"column:share_count;not null"`
	BookmarkCount         int       `gorm:"column:bookmark_count;not null"`
	UpdatedAt             time.Time `gorm:"column:updated_at;not null"`
}

func (gormPostMetrics) TableName() string { return "post_metrics" }

func (r *GormRepository) findByIDWithExtras(ctx context.Context, db *gorm.DB, id ident.PostID) (*model.Post, error) {
	var row gormPost
	if err := db.
		Where("id = ?", string(id)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	p := toDomain(&row)
	seo, _ := loadSEO(ctx, db, id)
	metrics, _ := loadMetrics(ctx, db, id)
	p.SEO = seo
	p.Metrics = metrics

	// Load user summaries.
	p.Author, _ = loadUserSummary(ctx, db, row.AuthorID)
	if row.ReviewerUserID != nil {
		p.Reviewer, _ = loadUserSummary(ctx, db, *row.ReviewerUserID)
	}
	if row.LastEditedByUserID != nil {
		p.LastEditedBy, _ = loadUserSummary(ctx, db, *row.LastEditedByUserID)
	}

	// Build editors list from reviewer + lastEditedBy.
	editorIDs := make([]string, 0, 2)
	if row.ReviewerUserID != nil {
		editorIDs = append(editorIDs, strconv.FormatInt(*row.ReviewerUserID, 10))
	}
	if row.LastEditedByUserID != nil {
		seen := false
		for _, id := range editorIDs {
			if id == strconv.FormatInt(*row.LastEditedByUserID, 10) {
				seen = true
				break
			}
		}
		if !seen {
			editorIDs = append(editorIDs, strconv.FormatInt(*row.LastEditedByUserID, 10))
		}
	}
	p.EditorUserIDs = editorIDs
	p.Editors = make([]model.UserSummary, 0, len(editorIDs))
	for _, eid := range editorIDs {
		if u, _ := loadUserSummary(ctx, db, parseIDString(eid)); u != nil {
			p.Editors = append(p.Editors, *u)
		}
	}

	// Load taxonomy.
	p.Categories, _ = loadPostCategories(ctx, db, string(id), row.PrimaryCategoryID)
	p.Tags, _ = loadPostTags(ctx, db, string(id))

	// Load series for this post (optional).
	p.Series, _ = loadPostSeries(ctx, db, string(id))

	// Load syndication/changelog/gallery/translations (optional).
	p.Syndication, _ = loadPostSyndication(ctx, db, string(id))
	p.Changelog, _ = loadPostChangelog(ctx, db, string(id))
	p.Gallery, _ = loadPostGallery(ctx, db, string(id))
	p.Translations, _ = loadPostTranslations(ctx, db, string(id))
	p.CoAuthors, _ = loadPostCoauthors(ctx, db, string(id))
	if row.FeaturedMediaID != nil && *row.FeaturedMediaID != "" {
		var url string
		if err := db.WithContext(ctx).Table("media").Select("public_url").Where("id = ?", *row.FeaturedMediaID).Scan(&url).Error; err == nil && url != "" {
			p.FeaturedMediaPublicURL = &url
		}
	}

	return p, nil
}

func loadPostCoauthors(ctx context.Context, db *gorm.DB, postID string) ([]model.UserSummary, error) {
	var rows []struct {
		UserID    int64  `gorm:"column:user_id"`
		SortOrder int    `gorm:"column:sort_order"`
	}
	if err := db.WithContext(ctx).
		Table("post_coauthors").
		Select("user_id, sort_order").
		Where("post_id = ?", postID).
		Order("sort_order ASC, user_id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.UserSummary, 0, len(rows))
	for _, r := range rows {
		if u, _ := loadUserSummary(ctx, db, r.UserID); u != nil {
			out = append(out, *u)
		}
	}
	return out, nil
}

func loadUserSummary(ctx context.Context, db *gorm.DB, userID int64) (*model.UserSummary, error) {
	if userID <= 0 {
		return nil, nil
	}
	var u gormUserSummary
	if err := db.WithContext(ctx).
		Table("users").
		Select("public_id, id AS uuid, first_name, last_name, email").
		Where("public_id = ?", userID).
		First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	email := u.Email
	display := strings.TrimSpace(strings.TrimSpace(u.FirstName) + " " + strings.TrimSpace(u.LastName))
	return &model.UserSummary{
		ID:          strconv.FormatInt(u.ID, 10),
		UUID:        u.UUID,
		DisplayName: display,
		Email:       &email,
	}, nil
}

type gormPostCategoryRow struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

func loadPostCategories(ctx context.Context, db *gorm.DB, postID string, primaryCategoryID *string) ([]model.PostCategory, error) {
	var rows []gormPostCategoryRow
	if err := db.WithContext(ctx).
		Table("categories c").
		Select("c.id, c.name, c.slug").
		Joins("JOIN post_categories pc ON pc.category_id = c.id").
		Where("pc.post_id = ?", postID).
		Order("c.name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostCategory, 0, len(rows))
	for _, r := range rows {
		isPrimary := primaryCategoryID != nil && *primaryCategoryID == r.ID
		out = append(out, model.PostCategory{ID: r.ID, Name: r.Name, Slug: r.Slug, IsPrimary: isPrimary})
	}
	return out, nil
}

type gormPostTagRow struct {
	ID   string `gorm:"column:id"`
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

func loadPostTags(ctx context.Context, db *gorm.DB, postID string) ([]model.PostTag, error) {
	var rows []gormPostTagRow
	if err := db.WithContext(ctx).
		Table("tags t").
		Select("t.id, t.name, t.slug").
		Joins("JOIN post_tags pt ON pt.tag_id = t.id").
		Where("pt.post_id = ?", postID).
		Order("t.name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostTag, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.PostTag{ID: r.ID, Name: r.Name, Slug: r.Slug})
	}
	return out, nil
}

type gormPostSeriesRow struct {
	ID        string  `gorm:"column:id"`
	Title     string  `gorm:"column:title"`
	Slug      string  `gorm:"column:slug"`
	PartIndex *int    `gorm:"column:part_index"`
	PartLabel *string `gorm:"column:part_label"`
}

func loadPostSeries(ctx context.Context, db *gorm.DB, postID string) (*model.PostSeries, error) {
	var row gormPostSeriesRow
	if err := db.WithContext(ctx).
		Table("post_series ps").
		Select("s.id, s.title, s.slug, ps.part_index, ps.part_label").
		Joins("JOIN series s ON s.id = ps.series_id").
		Where("ps.post_id = ?", postID).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &model.PostSeries{
		ID:        row.ID,
		Title:     row.Title,
		Slug:      row.Slug,
		PartIndex: row.PartIndex,
		PartLabel: row.PartLabel,
	}, nil
}

type gormPostSyndicationRow struct {
	ID       string `gorm:"column:id"`
	Platform string `gorm:"column:platform"`
	URL      string `gorm:"column:url"`
	Status   string `gorm:"column:status"`
}

func loadPostSyndication(ctx context.Context, db *gorm.DB, postID string) ([]model.PostSyndication, error) {
	var rows []gormPostSyndicationRow
	if err := db.WithContext(ctx).
		Table("post_syndication").
		Select("id, platform, url, status").
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostSyndication, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.PostSyndication{ID: r.ID, Platform: r.Platform, URL: r.URL, Status: r.Status})
	}
	return out, nil
}

type gormPostChangelogRow struct {
	ID     string    `gorm:"column:id"`
	At     time.Time `gorm:"column:at"`
	UserID *int64    `gorm:"column:user_id"`
	Note   string    `gorm:"column:note"`
}

func loadPostChangelog(ctx context.Context, db *gorm.DB, postID string) ([]model.PostChangelogEntry, error) {
	var rows []gormPostChangelogRow
	if err := db.WithContext(ctx).
		Table("post_changelog").
		Select("id, at, user_id, note").
		Where("post_id = ?", postID).
		Order("at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostChangelogEntry, 0, len(rows))
	for _, r := range rows {
		var u *model.UserSummary
		if r.UserID != nil {
			u, _ = loadUserSummary(ctx, db, *r.UserID)
		}
		out = append(out, model.PostChangelogEntry{ID: r.ID, At: r.At, User: u, Note: r.Note})
	}
	return out, nil
}

type gormPostGalleryRow struct {
	ID        string  `gorm:"column:id"`
	MediaID   string  `gorm:"column:media_id"`
	URL       string  `gorm:"column:public_url"`
	SortOrder int     `gorm:"column:sort_order"`
	Caption   *string `gorm:"column:caption"`
	Alt       *string `gorm:"column:alt"`
}

func loadPostGallery(ctx context.Context, db *gorm.DB, postID string) ([]model.PostGalleryItem, error) {
	var rows []gormPostGalleryRow
	if err := db.WithContext(ctx).
		Table("post_gallery_items gi").
		Select("gi.id, gi.media_id, m.public_url, gi.sort_order, gi.caption, gi.alt").
		Joins("JOIN media m ON m.id = gi.media_id").
		Where("gi.post_id = ?", postID).
		Order("gi.sort_order ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostGalleryItem, 0, len(rows))
	for _, r := range rows {
		url := r.URL
		out = append(out, model.PostGalleryItem{ID: r.ID, MediaID: r.MediaID, URL: &url, SortOrder: r.SortOrder, Caption: r.Caption, Alt: r.Alt})
	}
	return out, nil
}

type gormPostTranslationRow struct {
	PostID  string `gorm:"column:post_id"`
	Locale  string `gorm:"column:locale"`
	Slug    string `gorm:"column:slug"`
	GroupID string `gorm:"column:group_id"`
}

func loadPostTranslations(ctx context.Context, db *gorm.DB, postID string) (*model.PostTranslations, error) {
	// Find the group for this post (if any).
	var group struct {
		GroupID string `gorm:"column:group_id"`
		Locale  string `gorm:"column:locale"`
	}
	if err := db.WithContext(ctx).
		Table("post_translations").
		Select("group_id, locale").
		Where("post_id = ?", postID).
		First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.PostTranslations{GroupID: nil, Translations: []model.PostTranslationEntry{}}, nil
		}
		return nil, err
	}
	gid := group.GroupID
	var rows []gormPostTranslationRow
	if err := db.WithContext(ctx).
		Table("post_translations pt").
		Select("pt.post_id, pt.locale, pt.group_id, p.slug").
		Joins("JOIN posts p ON p.id = pt.post_id").
		Where("pt.group_id = ?", gid).
		Order("pt.locale ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostTranslationEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.PostTranslationEntry{PostID: r.PostID, Locale: r.Locale, Slug: r.Slug})
	}
	return &model.PostTranslations{GroupID: &gid, Translations: out}, nil
}

func loadSEO(ctx context.Context, db *gorm.DB, postID ident.PostID) (*seo.PostSEO, error) {
	var row gormPostSEO
	if err := db.WithContext(ctx).Where("post_id = ?", string(postID)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &seo.PostSEO{
		Title:          row.Title,
		Description:    row.Description,
		CanonicalURL:   row.CanonicalURL,
		Robots:         row.Robots,
		OGType:         row.OGType,
		OGImageURL:     row.OGImageURL,
		TwitterCard:    row.TwitterCard,
		StructuredData: row.StructuredData,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func loadMetrics(ctx context.Context, db *gorm.DB, postID ident.PostID) (*metrics.PostMetrics, error) {
	var row gormPostMetrics
	if err := db.WithContext(ctx).Where("post_id = ?", string(postID)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &metrics.PostMetrics{
		WordCount:             row.WordCount,
		CharacterCount:        row.CharacterCount,
		ReadingTimeMinutes:    row.ReadingTimeMinutes,
		EstReadTimeSeconds:    row.EstReadTimeSeconds,
		ViewCount:             row.ViewCount,
		UniqueVisitors7d:      row.UniqueVisitors7d,
		ScrollDepthAvgPercent: row.ScrollDepthAvgPercent,
		BounceRatePercent:     row.BounceRatePercent,
		AvgTimeOnPageSeconds:  row.AvgTimeOnPageSeconds,
		CommentCount:          row.CommentCount,
		LikeCount:             row.LikeCount,
		ShareCount:            row.ShareCount,
		BookmarkCount:         row.BookmarkCount,
		UpdatedAt:             row.UpdatedAt,
	}, nil
}

func upsertSEO(ctx context.Context, tx *gorm.DB, postID ident.PostID, seo *seo.PostSEO) error {
	if seo == nil {
		return nil
	}
	row := &gormPostSEO{
		PostID:         string(postID),
		Title:          seo.Title,
		Description:    seo.Description,
		CanonicalURL:   seo.CanonicalURL,
		Robots:         seo.Robots,
		OGType:         seo.OGType,
		OGImageURL:     seo.OGImageURL,
		TwitterCard:    seo.TwitterCard,
		StructuredData: nonNilJSON(seo.StructuredData),
		UpdatedAt:      time.Now().UTC(),
	}
	return tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "description", "canonical_url", "robots", "og_type", "og_image_url", "twitter_card", "structured_data", "updated_at"}),
		}).
		Create(row).Error
}

func upsertMetrics(ctx context.Context, tx *gorm.DB, postID ident.PostID, m *metrics.PostMetrics) error {
	if m == nil {
		return nil
	}
	row := &gormPostMetrics{
		PostID:                string(postID),
		WordCount:             m.WordCount,
		CharacterCount:        m.CharacterCount,
		ReadingTimeMinutes:    m.ReadingTimeMinutes,
		EstReadTimeSeconds:    m.EstReadTimeSeconds,
		ViewCount:             m.ViewCount,
		UniqueVisitors7d:      m.UniqueVisitors7d,
		ScrollDepthAvgPercent: m.ScrollDepthAvgPercent,
		BounceRatePercent:     m.BounceRatePercent,
		AvgTimeOnPageSeconds:  m.AvgTimeOnPageSeconds,
		CommentCount:          m.CommentCount,
		LikeCount:             m.LikeCount,
		ShareCount:            m.ShareCount,
		BookmarkCount:         m.BookmarkCount,
		UpdatedAt:             time.Now().UTC(),
	}
	return tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "post_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"word_count",
				"character_count",
				"reading_time_minutes",
				"est_read_time_seconds",
				"view_count",
				"unique_visitors_7d",
				"scroll_depth_avg_percent",
				"bounce_rate_percent",
				"avg_time_on_page_seconds",
				"comment_count",
				"like_count",
				"share_count",
				"bookmark_count",
				"updated_at",
			}),
		}).
		Create(row).Error
}

func nonNilJSON(b []byte) json.RawMessage {
	if len(b) == 0 {
		return json.RawMessage([]byte(`{}`))
	}
	return json.RawMessage(b)
}

// --- Sub-resource CRUD (admin) ---

func (r *GormRepository) DeleteSEO(ctx context.Context, postID ident.PostID) error {
	return r.db.WithContext(ctx).Where("post_id = ?", string(postID)).Delete(&gormPostSEO{}).Error
}

func (r *GormRepository) UpsertSEOOnly(ctx context.Context, postID ident.PostID, seo *seo.PostSEO) error {
	if seo == nil {
		return errors.New("seo_required")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return upsertSEO(ctx, tx, postID, seo)
	})
}

func (r *GormRepository) GetMetrics(ctx context.Context, postID ident.PostID) (*metrics.PostMetrics, error) {
	return loadMetrics(ctx, r.db.WithContext(ctx), postID)
}

func (r *GormRepository) SetFeaturedImage(ctx context.Context, postID ident.PostID, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error {
	updates := map[string]any{
		"featured_media_id": mediaID,
		"featured_alt":      alt,
		"featured_width":    width,
		"featured_height":   height,
		"featured_focal_x":  focalX,
		"featured_focal_y":  focalY,
		"featured_credit":   credit,
		"featured_license":  license,
		"updated_at":        time.Now().UTC(),
	}
	return r.db.WithContext(ctx).Model(&gormPost{}).Where("id = ?", string(postID)).Updates(updates).Error
}

type gormPostSeriesLink struct {
	PostID    string  `gorm:"column:post_id;type:uuid;primaryKey"`
	SeriesID  string  `gorm:"column:series_id;type:uuid;primaryKey"`
	PartIndex *int    `gorm:"column:part_index"`
	PartLabel *string `gorm:"column:part_label"`
}

func (gormPostSeriesLink) TableName() string { return "post_series" }

func (r *GormRepository) SetPostSeries(ctx context.Context, postID ident.PostID, seriesID *string, partIndex *int, partLabel *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostSeriesLink{}).Error; err != nil {
			return err
		}
		if seriesID == nil || strings.TrimSpace(*seriesID) == "" {
			return nil
		}
		sid := strings.TrimSpace(*seriesID)
		row := gormPostSeriesLink{
			PostID:    string(postID),
			SeriesID:  sid,
			PartIndex: partIndex,
			PartLabel: partLabel,
		}
		return tx.Create(&row).Error
	})
}

type gormPostCoauthor struct {
	PostID    string `gorm:"column:post_id;type:uuid;primaryKey"`
	UserID    string `gorm:"column:user_id;type:uuid;primaryKey"`
	SortOrder int    `gorm:"column:sort_order"`
}

func (gormPostCoauthor) TableName() string { return "post_coauthors" }

func (r *GormRepository) ReplaceCoauthors(ctx context.Context, postID ident.PostID, userIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostCoauthor{}).Error; err != nil {
			return err
		}
		for i, uid := range userIDs {
			uid = strings.TrimSpace(uid)
			if uid == "" {
				continue
			}
			if err := tx.Create(&gormPostCoauthor{PostID: string(postID), UserID: uid, SortOrder: i}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type gormPostGalleryItemRow struct {
	ID        string  `gorm:"column:id;type:uuid;primaryKey"`
	PostID    string  `gorm:"column:post_id;type:uuid;not null"`
	MediaID   string  `gorm:"column:media_id;type:uuid;not null"`
	SortOrder int     `gorm:"column:sort_order;not null"`
	Caption   *string `gorm:"column:caption"`
	Alt       *string `gorm:"column:alt"`
}

func (gormPostGalleryItemRow) TableName() string { return "post_gallery_items" }

func (r *GormRepository) CreateGalleryItem(ctx context.Context, postID ident.PostID, mediaID string, sortOrder int, caption *string, alt *string) (string, error) {
	mediaID = strings.TrimSpace(mediaID)
	if mediaID == "" {
		return "", errors.New("media_id_required")
	}
	id := uuid.NewString()
	row := &gormPostGalleryItemRow{
		ID:        id,
		PostID:    string(postID),
		MediaID:   mediaID,
		SortOrder: sortOrder,
		Caption:   caption,
		Alt:       alt,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return "", err
	}
	return id, nil
}

func (r *GormRepository) UpdateGalleryItem(ctx context.Context, postID ident.PostID, itemID string, sortOrder *int, caption *string, alt *string) error {
	updates := map[string]any{}
	if sortOrder != nil {
		updates["sort_order"] = *sortOrder
	}
	if caption != nil {
		updates["caption"] = caption
	}
	if alt != nil {
		updates["alt"] = alt
	}
	res := r.db.WithContext(ctx).Model(&gormPostGalleryItemRow{}).
		Where("id = ? AND post_id = ?", strings.TrimSpace(itemID), string(postID)).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) DeleteGalleryItem(ctx context.Context, postID ident.PostID, itemID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND post_id = ?", strings.TrimSpace(itemID), string(postID)).
		Delete(&gormPostGalleryItemRow{}).Error
}

type gormPostChangelog struct {
	ID     string    `gorm:"column:id;type:uuid;primaryKey"`
	PostID string    `gorm:"column:post_id;type:uuid;not null"`
	At     time.Time `gorm:"column:at;not null"`
	UserID *string   `gorm:"column:user_id;type:uuid"`
	Note   string    `gorm:"column:note;not null"`
}

func (gormPostChangelog) TableName() string { return "post_changelog" }

func (r *GormRepository) CreateChangelog(ctx context.Context, postID ident.PostID, userID *string, note string) (string, error) {
	note = strings.TrimSpace(note)
	if note == "" {
		return "", errors.New("note_required")
	}
	id := uuid.NewString()
	row := gormPostChangelog{
		ID:     id,
		PostID: string(postID),
		At:     time.Now().UTC(),
		UserID: userID,
		Note:   note,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", err
	}
	return id, nil
}

func (r *GormRepository) DeleteChangelog(ctx context.Context, postID ident.PostID, changelogID string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND post_id = ?", strings.TrimSpace(changelogID), string(postID)).
		Delete(&gormPostChangelog{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type gormPostSyndicationRowFull struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	PostID    string    `gorm:"column:post_id;type:uuid;not null"`
	Platform  string    `gorm:"column:platform;not null"`
	URL       string    `gorm:"column:url;not null"`
	Status    string    `gorm:"column:status;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (gormPostSyndicationRowFull) TableName() string { return "post_syndication" }

func (r *GormRepository) CreateSyndication(ctx context.Context, postID ident.PostID, platform, url, status string) (string, error) {
	platform = strings.TrimSpace(platform)
	url = strings.TrimSpace(url)
	if platform == "" || url == "" {
		return "", errors.New("platform_url_required")
	}
	if strings.TrimSpace(status) == "" {
		status = "active"
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	row := gormPostSyndicationRowFull{
		ID:        id,
		PostID:    string(postID),
		Platform:  platform,
		URL:       url,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", err
	}
	return id, nil
}

func (r *GormRepository) UpdateSyndication(ctx context.Context, postID ident.PostID, id string, platform, url, status *string) error {
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if platform != nil {
		updates["platform"] = strings.TrimSpace(*platform)
	}
	if url != nil {
		updates["url"] = strings.TrimSpace(*url)
	}
	if status != nil {
		updates["status"] = strings.TrimSpace(*status)
	}
	res := r.db.WithContext(ctx).Model(&gormPostSyndicationRowFull{}).
		Where("id = ? AND post_id = ?", strings.TrimSpace(id), string(postID)).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) DeleteSyndication(ctx context.Context, postID ident.PostID, id string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND post_id = ?", strings.TrimSpace(id), string(postID)).
		Delete(&gormPostSyndicationRowFull{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) UpdateSyndicationByID(ctx context.Context, id string, platform, url, status *string) error {
	updates := map[string]any{"updated_at": time.Now().UTC()}
	if platform != nil {
		updates["platform"] = strings.TrimSpace(*platform)
	}
	if url != nil {
		updates["url"] = strings.TrimSpace(*url)
	}
	if status != nil {
		updates["status"] = strings.TrimSpace(*status)
	}
	res := r.db.WithContext(ctx).Model(&gormPostSyndicationRowFull{}).
		Where("id = ?", strings.TrimSpace(id)).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) DeleteSyndicationByID(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Where("id = ?", strings.TrimSpace(id)).
		Delete(&gormPostSyndicationRowFull{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

type gormTranslationGroupRow struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (gormTranslationGroupRow) TableName() string { return "translation_groups" }

type gormPostTranslationRowFull struct {
	PostID  string `gorm:"column:post_id;type:uuid;primaryKey"`
	GroupID string `gorm:"column:group_id;type:uuid;not null"`
	Locale  string `gorm:"column:locale;not null"`
}

func (gormPostTranslationRowFull) TableName() string { return "post_translations" }

func (r *GormRepository) PutPostTranslation(ctx context.Context, postID ident.PostID, groupID *string, locale string) (string, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return "", errors.New("locale_required")
	}
	var resolved string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostTranslationRowFull{}).Error; err != nil {
			return err
		}
		gid := ""
		if groupID != nil && strings.TrimSpace(*groupID) != "" {
			gid = strings.TrimSpace(*groupID)
			var n int64
			if err := tx.Model(&gormTranslationGroupRow{}).Where("id = ?", gid).Count(&n).Error; err != nil {
				return err
			}
			if n == 0 {
				return gorm.ErrRecordNotFound
			}
		} else {
			gid = uuid.NewString()
			now := time.Now().UTC()
			if err := tx.Create(&gormTranslationGroupRow{ID: gid, CreatedAt: now}).Error; err != nil {
				return err
			}
		}
		row := gormPostTranslationRowFull{PostID: string(postID), GroupID: gid, Locale: locale}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		resolved = gid
		return nil
	})
	return resolved, err
}

func (r *GormRepository) ClearPostTranslation(ctx context.Context, postID ident.PostID) error {
	return r.db.WithContext(ctx).Where("post_id = ?", string(postID)).Delete(&gormPostTranslationRowFull{}).Error
}

type gormSeriesRow struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Title     string    `gorm:"column:title;not null"`
	Slug      string    `gorm:"column:slug;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (gormSeriesRow) TableName() string { return "series" }

func (r *GormRepository) ListSeries(ctx context.Context) ([]series.Series, error) {
	var rows []gormSeriesRow
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]series.Series, 0, len(rows))
	for _, row := range rows {
		out = append(out, series.Series{ID: row.ID, Title: row.Title, Slug: row.Slug, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
	}
	return out, nil
}

func (r *GormRepository) CreateSeries(ctx context.Context, s *series.Series) error {
	if s == nil || strings.TrimSpace(s.Title) == "" || strings.TrimSpace(s.Slug) == "" {
		return errors.New("invalid_series")
	}
	now := time.Now().UTC()
	if strings.TrimSpace(s.ID) == "" {
		s.ID = uuid.NewString()
	}
	s.CreatedAt = now
	s.UpdatedAt = now
	row := gormSeriesRow{ID: s.ID, Title: strings.TrimSpace(s.Title), Slug: strings.TrimSpace(s.Slug), CreatedAt: now, UpdatedAt: now}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) FindSeriesByID(ctx context.Context, id string) (*series.Series, error) {
	var row gormSeriesRow
	if err := r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &series.Series{ID: row.ID, Title: row.Title, Slug: row.Slug, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (r *GormRepository) UpdateSeries(ctx context.Context, s *series.Series) error {
	if s == nil || strings.TrimSpace(s.ID) == "" {
		return errors.New("invalid_series")
	}
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&gormSeriesRow{}).
		Where("id = ?", strings.TrimSpace(s.ID)).
		Updates(map[string]any{
			"title":      strings.TrimSpace(s.Title),
			"slug":       strings.TrimSpace(s.Slug),
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) DeleteSeries(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(id)).Delete(&gormSeriesRow{}).Error
}

func (r *GormRepository) CreateTranslationGroup(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		id = uuid.NewString()
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Create(&gormTranslationGroupRow{ID: id, CreatedAt: now}).Error
}

func (r *GormRepository) FindTranslationGroup(ctx context.Context, id string) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&gormTranslationGroupRow{}).Where("id = ?", strings.TrimSpace(id)).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *GormRepository) DeleteTranslationGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(id)).Delete(&gormTranslationGroupRow{}).Error
}

func parseIDString(raw string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseOptionalIDString(raw *string) *int64 {
	if raw == nil {
		return nil
	}
	v := parseIDString(*raw)
	if v == 0 {
		return nil
	}
	return &v
}

func int64PtrToStringPtr(v *int64) *string {
	if v == nil || *v == 0 {
		return nil
	}
	s := strconv.FormatInt(*v, 10)
	return &s
}

var _ ports.Repository = (*GormRepository)(nil)
