package graphql

import (
	authApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/auth/application"
	pagesApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/application"
	menuApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/application"
	postApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/application"
	taxApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/taxonomy/application"
	platformES "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/elasticsearch"
)

// Resolver is the root GraphQL resolver; field resolvers live in generated companion files.
// PostsCore is named to avoid clashing with the generated Query.posts field resolver method.
type Resolver struct {
	Auth      *authApp.Service
	PostsCore *postApp.CorePostsService
	Pages     *pagesApp.Service
	Taxonomy  *taxApp.Service
	Menus     *menuApp.Service
	Search    *platformES.PostsIndex
}
