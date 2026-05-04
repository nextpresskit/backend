package graphql

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
