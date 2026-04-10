package graphql

import (
	pagesApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/application"
	postApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/application"
)

// Resolver is the root GraphQL resolver; field resolvers live in generated companion files.
// PostsCore is named to avoid clashing with the generated Query.posts field resolver method.
type Resolver struct {
	PostsCore *postApp.CorePostsService
	Pages     *pagesApp.Service
}
