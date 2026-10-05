package infrastructure

import (
	"context"
	"strings"

	"github.com/nextpresskit/backend/internal/modules/user/domain"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
)

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// List returns one page of non-deleted users ordered by id, plus the total number of matches.
func (r *GormRepository) List(ctx context.Context, f domain.ListFilter) ([]domain.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&userp.User{})
	if term := strings.TrimSpace(f.Query); term != "" {
		like := "%" + likeEscaper.Replace(strings.ToLower(term)) + "%"
		q = q.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(first_name || ' ' || last_name) LIKE ?",
			like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []userp.User
	if err := q.Order("id ASC").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.User, 0, len(rows))
	for i := range rows {
		out = append(out, *toDomain(&rows[i]))
	}
	return out, total, nil
}
