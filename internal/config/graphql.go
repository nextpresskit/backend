package config

// GraphQLConfig toggles the GraphQL HTTP endpoint (gqlgen).
type GraphQLConfig struct {
	Enabled            bool
	Path               string
	PlaygroundEnabled  bool
}

// LoadGraphQLConfig reads GRAPHQL_* environment variables.
func LoadGraphQLConfig() GraphQLConfig {
	return GraphQLConfig{
		Enabled:           parseBool(GetEnv("GRAPHQL_ENABLED", "false")),
		Path:              GetEnv("GRAPHQL_PATH", "/v1/graphql"),
		PlaygroundEnabled: parseBool(GetEnv("GRAPHQL_PLAYGROUND_ENABLED", "false")),
	}
}
