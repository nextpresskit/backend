package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
	postp "github.com/nextpresskit/backend/internal/modules/posts/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormUserSummary struct {
	ID        int64  `gorm:"column:id"`
	UUID      string `gorm:"column:uuid"`
	FirstName string `gorm:"column:first_name"`
	LastName  string `gorm:"column:last_name"`
	Email     string `gorm:"column:email"`
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := fromDomain(post)
		enrichPersistencePost(tx, m, post)
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

func (r *GormRepository) FindByUUID(ctx context.Context, u string) (*model.Post, error) {
	u = strings.TrimSpace(u)
	if u == "" {
		return nil, nil
	}
	var row postp.Post
	if err := r.db.WithContext(ctx).Where("uuid = ?", u).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.findByIDWithExtras(ctx, r.db.WithContext(ctx), ident.PostID(row.ID))
}

func (r *GormRepository) FindBySlug(ctx context.Context, slug string) (*model.Post, error) {
	var row postp.Post
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
	dbq := r.db.WithContext(ctx).Model(&postp.Post{}).Order("created_at DESC").Limit(limit).Offset(offset)
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

	var rows []postp.Post
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
		Model(&postp.Post{}).
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
			Joins("JOIN categories c ON c.id = pc.category_id").
			Where("c.uuid = ?", categoryID)
	}
	if tagID != "" {
		dbq = dbq.Joins("JOIN post_tags pt ON pt.post_id = posts.id").
			Joins("JOIN tags t ON t.id = pt.tag_id").
			Where("t.uuid = ?", tagID)
	}

	var rows []postp.Post
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
	var row postp.Post
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := fromDomain(post)
		enrichPersistencePost(tx, m, post)
		if err := tx.Model(&postp.Post{}).
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
		Where("id = ?", int64(id)).
		Delete(&postp.Post{}).Error
}

func (r *GormRepository) SetCategories(ctx context.Context, postID ident.PostID, categoryIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", int64(postID)).Delete(&postp.PostCategory{}).Error; err != nil {
			return err
		}
		if len(categoryIDs) == 0 {
			return nil
		}

		rows := make([]postp.PostCategory, 0, len(categoryIDs))
		for _, id := range categoryIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			var cid int64
			if err := tx.Table("categories").Select("id").Where("uuid = ?", id).Scan(&cid).Error; err != nil || cid == 0 {
				continue
			}
			rows = append(rows, postp.PostCategory{PostID: int64(postID), CategoryID: cid})
		}
		if len(rows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func (r *GormRepository) SetTags(ctx context.Context, postID ident.PostID, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", int64(postID)).Delete(&postp.PostTag{}).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}

		rows := make([]postp.PostTag, 0, len(tagIDs))
		for _, id := range tagIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			var tid int64
			if err := tx.Table("tags").Select("id").Where("uuid = ?", id).Scan(&tid).Error; err != nil || tid == 0 {
				continue
			}
			rows = append(rows, postp.PostTag{PostID: int64(postID), TagID: tid})
		}
		if len(rows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func (r *GormRepository) SetPrimaryCategory(ctx context.Context, postID ident.PostID, categoryID *string) error {
	var catID *int64
	if categoryID != nil {
		s := strings.TrimSpace(*categoryID)
		if s != "" {
			var cid int64
			if err := r.db.Table("categories").Select("id").Where("uuid = ?", s).Scan(&cid).Error; err == nil && cid > 0 {
				catID = &cid
			}
		}
	}
	return r.db.WithContext(ctx).
		Model(&postp.Post{}).
		Where("id = ?", int64(postID)).
		Update("primary_category_id", catID).Error
}

func toDomain(m *postp.Post) *model.Post {
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
		FeaturedMediaID:    nil,
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

func fromDomain(p *model.Post) *postp.Post {
	var deleted gorm.DeletedAt
	if p.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *p.DeletedAt, Valid: true}
	}
	return &postp.Post{
		ID:                 int64(p.ID),
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
		FeaturedMediaID:    nil,
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

// enrichPersistencePost maps public media/category UUIDs on the domain model to FK ints for storage.
func enrichPersistencePost(tx *gorm.DB, m *postp.Post, p *model.Post) {
	if p.FeaturedMediaID != nil {
		s := strings.TrimSpace(*p.FeaturedMediaID)
		if s != "" {
			var mid int64
			_ = tx.Table("media").Select("id").Where("uuid = ?", s).Scan(&mid).Error
			if mid > 0 {
				m.FeaturedMediaID = &mid
			}
		} else {
			m.FeaturedMediaID = nil
		}
	} else {
		m.FeaturedMediaID = nil
	}
}

func (r *GormRepository) findByIDWithExtras(ctx context.Context, db *gorm.DB, id ident.PostID) (*model.Post, error) {
	var row postp.Post
	if err := db.
		Where("id = ?", int64(id)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	p := toDomain(&row)
	if row.FeaturedMediaID != nil {
		var mu string
		if err := db.WithContext(ctx).Table("media").Select("uuid").Where("id = ?", *row.FeaturedMediaID).Scan(&mu).Error; err == nil && mu != "" {
			p.FeaturedMediaID = &mu
		}
	}
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
	p.Categories, _ = loadPostCategories(ctx, db, int64(id), row.PrimaryCategoryID)
	p.Tags, _ = loadPostTags(ctx, db, int64(id))

	// Load series for this post (optional).
	p.Series, _ = loadPostSeries(ctx, db, int64(id))

	// Load syndication/changelog/gallery/translations (optional).
	p.Syndication, _ = loadPostSyndication(ctx, db, int64(id))
	p.Changelog, _ = loadPostChangelog(ctx, db, int64(id))
	p.Gallery, _ = loadPostGallery(ctx, db, int64(id))
	p.Translations, _ = loadPostTranslations(ctx, db, int64(id))
	p.CoAuthors, _ = loadPostCoauthors(ctx, db, int64(id))
	if row.FeaturedMediaID != nil && *row.FeaturedMediaID != 0 {
		var url string
		if err := db.WithContext(ctx).Table("media").Select("public_url").Where("id = ?", *row.FeaturedMediaID).Scan(&url).Error; err == nil && url != "" {
			p.FeaturedMediaPublicURL = &url
		}
	}

	return p, nil
}

func loadPostCoauthors(ctx context.Context, db *gorm.DB, postID int64) ([]model.UserSummary, error) {
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
		Select("id, uuid, first_name, last_name, email").
		Where("id = ?", userID).
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

type postCategoryJoinRow struct {
	InternalID int64  `gorm:"column:id"`
	UUID       string `gorm:"column:uuid"`
	Name       string `gorm:"column:name"`
	Slug       string `gorm:"column:slug"`
}

func loadPostCategories(ctx context.Context, db *gorm.DB, postID int64, primaryCategoryID *int64) ([]model.PostCategory, error) {
	var rows []postCategoryJoinRow
	if err := db.WithContext(ctx).
		Table("categories c").
		Select("c.id, c.uuid AS uuid, c.name, c.slug").
		Joins("JOIN post_categories pc ON pc.category_id = c.id").
		Where("pc.post_id = ?", postID).
		Order("c.name ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostCategory, 0, len(rows))
	for _, r := range rows {
		isPrimary := primaryCategoryID != nil && *primaryCategoryID == r.InternalID
		out = append(out, model.PostCategory{ID: r.UUID, Name: r.Name, Slug: r.Slug, IsPrimary: isPrimary})
	}
	return out, nil
}

type postTagJoinRow struct {
	ID   string `gorm:"column:uuid"`
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

func loadPostTags(ctx context.Context, db *gorm.DB, postID int64) ([]model.PostTag, error) {
	var rows []postTagJoinRow
	if err := db.WithContext(ctx).
		Table("tags t").
		Select("t.uuid, t.name, t.slug").
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

type postSeriesJoinRow struct {
	ID        string  `gorm:"column:uuid"`
	Title     string  `gorm:"column:title"`
	Slug      string  `gorm:"column:slug"`
	PartIndex *int    `gorm:"column:part_index"`
	PartLabel *string `gorm:"column:part_label"`
}

func loadPostSeries(ctx context.Context, db *gorm.DB, postID int64) (*model.PostSeries, error) {
	var row postSeriesJoinRow
	if err := db.WithContext(ctx).
		Table("post_series ps").
		Select("s.uuid, s.title, s.slug, ps.part_index, ps.part_label").
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

type postSyndicationJoinRow struct {
	ID       string `gorm:"column:uuid"`
	Platform string `gorm:"column:platform"`
	URL      string `gorm:"column:url"`
	Status   string `gorm:"column:status"`
}

func loadPostSyndication(ctx context.Context, db *gorm.DB, postID int64) ([]model.PostSyndication, error) {
	var rows []postSyndicationJoinRow
	if err := db.WithContext(ctx).
		Table("post_syndication").
		Select("uuid, platform, url, status").
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

type postChangelogJoinRow struct {
	ID     string    `gorm:"column:uuid"`
	At     time.Time `gorm:"column:at"`
	UserID *int64    `gorm:"column:user_id"`
	Note   string    `gorm:"column:note"`
}

func loadPostChangelog(ctx context.Context, db *gorm.DB, postID int64) ([]model.PostChangelogEntry, error) {
	var rows []postChangelogJoinRow
	if err := db.WithContext(ctx).
		Table("post_changelog").
		Select("uuid, at, user_id, note").
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

type postGalleryJoinRow struct {
	ID        string  `gorm:"column:gi_uuid"`
	MediaID   string  `gorm:"column:m_uuid"`
	URL       string  `gorm:"column:public_url"`
	SortOrder int     `gorm:"column:sort_order"`
	Caption   *string `gorm:"column:caption"`
	Alt       *string `gorm:"column:alt"`
}

func loadPostGallery(ctx context.Context, db *gorm.DB, postID int64) ([]model.PostGalleryItem, error) {
	var rows []postGalleryJoinRow
	if err := db.WithContext(ctx).
		Table("post_gallery_items gi").
		Select("gi.uuid AS gi_uuid, m.uuid AS m_uuid, m.public_url, gi.sort_order, gi.caption, gi.alt").
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

type postTranslationJoinRow struct {
	PostUUID  string `gorm:"column:post_uuid"`
	Locale    string `gorm:"column:locale"`
	Slug      string `gorm:"column:slug"`
	GroupUUID string `gorm:"column:group_uuid"`
}

func loadPostTranslations(ctx context.Context, db *gorm.DB, postID int64) (*model.PostTranslations, error) {
	var group struct {
		GroupID int64  `gorm:"column:group_id"`
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
	var groupUUID string
	_ = db.WithContext(ctx).Table("translation_groups").Select("uuid").Where("id = ?", gid).Scan(&groupUUID).Error
	var rows []postTranslationJoinRow
	if err := db.WithContext(ctx).
		Table("post_translations pt").
		Select("p.uuid AS post_uuid, pt.locale, tg.uuid AS group_uuid, p.slug").
		Joins("JOIN posts p ON p.id = pt.post_id").
		Joins("JOIN translation_groups tg ON tg.id = pt.group_id").
		Where("pt.group_id = ?", gid).
		Order("pt.locale ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.PostTranslationEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.PostTranslationEntry{PostID: r.PostUUID, Locale: r.Locale, Slug: r.Slug})
	}
	gu := groupUUID
	return &model.PostTranslations{GroupID: &gu, Translations: out}, nil
}

func loadSEO(ctx context.Context, db *gorm.DB, postID ident.PostID) (*seo.PostSEO, error) {
	var row postp.PostSEO
	if err := db.WithContext(ctx).Where("post_id = ?", int64(postID)).First(&row).Error; err != nil {
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
	var row postp.PostMetrics
	if err := db.WithContext(ctx).Where("post_id = ?", int64(postID)).First(&row).Error; err != nil {
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
	row := &postp.PostSEO{
		PostID:         int64(postID),
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
	row := &postp.PostMetrics{
		PostID:                int64(postID),
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
	return r.db.WithContext(ctx).Where("post_id = ?", int64(postID)).Delete(&postp.PostSEO{}).Error
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
		"featured_alt":     alt,
		"featured_width":   width,
		"featured_height":  height,
		"featured_focal_x": focalX,
		"featured_focal_y": focalY,
		"featured_credit":  credit,
		"featured_license": license,
		"updated_at":       time.Now().UTC(),
	}
	if mediaID != nil {
		s := strings.TrimSpace(*mediaID)
		if s != "" {
			var mid int64
			_ = r.db.WithContext(ctx).Table("media").Select("id").Where("uuid = ?", s).Scan(&mid).Error
			if mid > 0 {
				updates["featured_media_id"] = mid
			}
		} else {
			updates["featured_media_id"] = nil
		}
	}
	return r.db.WithContext(ctx).Model(&postp.Post{}).Where("id = ?", int64(postID)).Updates(updates).Error
}

func (r *GormRepository) SetPostSeries(ctx context.Context, postID ident.PostID, seriesID *string, partIndex *int, partLabel *string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", int64(postID)).Delete(&postp.PostSeries{}).Error; err != nil {
			return err
		}
		if seriesID == nil || strings.TrimSpace(*seriesID) == "" {
			return nil
		}
		sidStr := strings.TrimSpace(*seriesID)
		var sid int64
		if err := tx.Table("series").Select("id").Where("uuid = ?", sidStr).Scan(&sid).Error; err != nil || sid == 0 {
			return gorm.ErrRecordNotFound
		}
		row := postp.PostSeries{
			PostID:    int64(postID),
			SeriesID:  sid,
			PartIndex: partIndex,
			PartLabel: partLabel,
		}
		return tx.Create(&row).Error
	})
}

func (r *GormRepository) ReplaceCoauthors(ctx context.Context, postID ident.PostID, userIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", int64(postID)).Delete(&postp.PostCoauthor{}).Error; err != nil {
			return err
		}
		for i, uid := range userIDs {
			uid = strings.TrimSpace(uid)
			if uid == "" {
				continue
			}
			pid, err := strconv.ParseInt(uid, 10, 64)
			if err != nil || pid <= 0 {
				continue
			}
			if err := tx.Create(&postp.PostCoauthor{PostID: int64(postID), UserID: pid, SortOrder: i}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormRepository) CreateGalleryItem(ctx context.Context, postID ident.PostID, mediaID string, sortOrder int, caption *string, alt *string) (string, error) {
	mediaID = strings.TrimSpace(mediaID)
	if mediaID == "" {
		return "", errors.New("media_id_required")
	}
	var mid int64
	if err := r.db.WithContext(ctx).Table("media").Select("id").Where("uuid = ?", mediaID).Scan(&mid).Error; err != nil || mid == 0 {
		return "", errors.New("media_not_found")
	}
	uid := uuid.NewString()
	row := &postp.PostGalleryItem{
		UUID:      uid,
		PostID:    int64(postID),
		MediaID:   mid,
		SortOrder: sortOrder,
		Caption:   caption,
		Alt:       alt,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return "", err
	}
	return uid, nil
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
	res := r.db.WithContext(ctx).Model(&postp.PostGalleryItem{}).
		Where("uuid = ? AND post_id = ?", strings.TrimSpace(itemID), int64(postID)).
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
		Where("uuid = ? AND post_id = ?", strings.TrimSpace(itemID), int64(postID)).
		Delete(&postp.PostGalleryItem{}).Error
}

func (r *GormRepository) CreateChangelog(ctx context.Context, postID ident.PostID, userID *string, note string) (string, error) {
	note = strings.TrimSpace(note)
	if note == "" {
		return "", errors.New("note_required")
	}
	var userPublic *int64
	if userID != nil {
		v := strings.TrimSpace(*userID)
		if v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
				userPublic = &n
			}
		}
	}
	cid := uuid.NewString()
	row := postp.PostChangelog{
		UUID:   cid,
		PostID: int64(postID),
		At:     time.Now().UTC(),
		UserID: userPublic,
		Note:   note,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", err
	}
	return cid, nil
}

func (r *GormRepository) DeleteChangelog(ctx context.Context, postID ident.PostID, changelogID string) error {
	res := r.db.WithContext(ctx).
		Where("uuid = ? AND post_id = ?", strings.TrimSpace(changelogID), int64(postID)).
		Delete(&postp.PostChangelog{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) CreateSyndication(ctx context.Context, postID ident.PostID, platform, url, status string) (string, error) {
	platform = strings.TrimSpace(platform)
	url = strings.TrimSpace(url)
	if platform == "" || url == "" {
		return "", errors.New("platform_url_required")
	}
	if strings.TrimSpace(status) == "" {
		status = "active"
	}
	sid := uuid.NewString()
	now := time.Now().UTC()
	row := postp.PostSyndication{
		UUID:      sid,
		PostID:    int64(postID),
		Platform:  platform,
		URL:       url,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", err
	}
	return sid, nil
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
	res := r.db.WithContext(ctx).Model(&postp.PostSyndication{}).
		Where("uuid = ? AND post_id = ?", strings.TrimSpace(id), int64(postID)).
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
		Where("uuid = ? AND post_id = ?", strings.TrimSpace(id), int64(postID)).
		Delete(&postp.PostSyndication{})
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
	res := r.db.WithContext(ctx).Model(&postp.PostSyndication{}).
		Where("uuid = ?", strings.TrimSpace(id)).
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
		Where("uuid = ?", strings.TrimSpace(id)).
		Delete(&postp.PostSyndication{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormRepository) PutPostTranslation(ctx context.Context, postID ident.PostID, groupID *string, locale string) (string, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return "", errors.New("locale_required")
	}
	var resolved string
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", int64(postID)).Delete(&postp.PostTranslation{}).Error; err != nil {
			return err
		}
		var gid int64
		if groupID != nil && strings.TrimSpace(*groupID) != "" {
			gu := strings.TrimSpace(*groupID)
			if err := tx.Table("translation_groups").Select("id").Where("uuid = ?", gu).Scan(&gid).Error; err != nil {
				return err
			}
			if gid == 0 {
				return gorm.ErrRecordNotFound
			}
			resolved = gu
		} else {
			nu := uuid.NewString()
			now := time.Now().UTC()
			tg := postp.TranslationGroup{UUID: nu, CreatedAt: now}
			if err := tx.Create(&tg).Error; err != nil {
				return err
			}
			gid = tg.ID
			resolved = nu
		}
		row := postp.PostTranslation{PostID: int64(postID), GroupID: gid, Locale: locale}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return nil
	})
	return resolved, err
}

func (r *GormRepository) ClearPostTranslation(ctx context.Context, postID ident.PostID) error {
	return r.db.WithContext(ctx).Where("post_id = ?", int64(postID)).Delete(&postp.PostTranslation{}).Error
}

func (r *GormRepository) ListSeries(ctx context.Context) ([]series.Series, error) {
	var rows []postp.Series
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]series.Series, 0, len(rows))
	for _, row := range rows {
		out = append(out, series.Series{ID: row.UUID, Title: row.Title, Slug: row.Slug, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
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
	row := postp.Series{UUID: s.ID, Title: strings.TrimSpace(s.Title), Slug: strings.TrimSpace(s.Slug), CreatedAt: now, UpdatedAt: now}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) FindSeriesByID(ctx context.Context, id string) (*series.Series, error) {
	var row postp.Series
	if err := r.db.WithContext(ctx).Where("uuid = ?", strings.TrimSpace(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &series.Series{ID: row.UUID, Title: row.Title, Slug: row.Slug, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (r *GormRepository) UpdateSeries(ctx context.Context, s *series.Series) error {
	if s == nil || strings.TrimSpace(s.ID) == "" {
		return errors.New("invalid_series")
	}
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&postp.Series{}).
		Where("uuid = ?", strings.TrimSpace(s.ID)).
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
	return r.db.WithContext(ctx).Where("uuid = ?", strings.TrimSpace(id)).Delete(&postp.Series{}).Error
}

func (r *GormRepository) CreateTranslationGroup(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		id = uuid.NewString()
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Create(&postp.TranslationGroup{UUID: id, CreatedAt: now}).Error
}

func (r *GormRepository) FindTranslationGroup(ctx context.Context, id string) (bool, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&postp.TranslationGroup{}).Where("uuid = ?", strings.TrimSpace(id)).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *GormRepository) DeleteTranslationGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("uuid = ?", strings.TrimSpace(id)).Delete(&postp.TranslationGroup{}).Error
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
