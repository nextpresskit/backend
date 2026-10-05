package domain

import "context"

// ListFilter narrows an admin user listing. Query matches name or email (case-insensitive).
type ListFilter struct {
	Query  string
	Limit  int
	Offset int
}

// Lister returns one page of non-deleted users plus the total match count.
type Lister interface {
	List(ctx context.Context, f ListFilter) ([]User, int64, error)
}

// RoleReader resolves role names for a batch of internal user ids (users.id).
// Implemented by the RBAC module; nil means "no RBAC wired" and yields empty role lists.
type RoleReader interface {
	RoleNamesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]string, error)
}
