package ports

import (
	"github.com/nextpresskit/backend/internal/modules/posts/domain/relations"
)

// CorePostsPersistence is the subset of Repository used by core post CRUD and taxonomy assignment.
type CorePostsPersistence interface {
	PostReader
	PostWriter
	relations.PostTaxonomyWriter
}
