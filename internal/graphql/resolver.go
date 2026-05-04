package graphql

import (
	authApp "github.com/nextpresskit/backend/internal/modules/auth/application"
	pagesApp "github.com/nextpresskit/backend/internal/modules/pages/application"
	postApp "github.com/nextpresskit/backend/internal/modules/posts/application"
	taxApp "github.com/nextpresskit/backend/internal/modules/taxonomy/application"
	platformES "github.com/nextpresskit/backend/internal/platform/elasticsearch"

	"github.com/nextpresskit/backend/internal/config"
)

// Resolver is the root GraphQL resolver; field resolvers live in generated companion files.
// PostsCore is named to avoid clashing with the generated Query.posts field resolver method.
type Resolver struct {
	Auth      *authApp.Service
	PostsCore *postApp.CorePostsService
	Pages     *pagesApp.Service
	Taxonomy  *taxApp.Service
	Search    *platformES.PostsIndex

	JWT config.JWTConfig
}
