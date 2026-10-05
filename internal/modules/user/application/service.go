package application

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/nextpresskit/backend/internal/modules/user/domain"
)

const (
	// DefaultListLimit is used when the caller sends no limit (or 0).
	DefaultListLimit = 20
	// MaxListLimit caps one page.
	MaxListLimit = 100
	// MaxQueryLength caps the search string.
	MaxQueryLength = 100
)

var (
	ErrUserNotFound   = errors.New("user_not_found")
	ErrInvalidUserID  = errors.New("invalid_user_id")
	ErrInvalidListArg = errors.New("invalid_pagination")
)

// UserWithRoles is a user plus its RBAC role names (never the password hash outside this package's callers).
type UserWithRoles struct {
	User  domain.User
	Roles []string
}

// ListResult is one page of users.
type ListResult struct {
	Users  []UserWithRoles
	Total  int64
	Limit  int
	Offset int
}

type Service struct {
	repo   domain.Repository
	lister domain.Lister
	roles  domain.RoleReader
}

// NewService wires the admin user read service. lister and roles may be nil (List then fails / roles stay empty).
func NewService(repo domain.Repository, lister domain.Lister, roles domain.RoleReader) *Service {
	return &Service{repo: repo, lister: lister, roles: roles}
}

// List returns a page of users. limit 0 means DefaultListLimit; limit is capped at MaxListLimit.
func (s *Service) List(ctx context.Context, query string, limit, offset int) (ListResult, error) {
	if limit < 0 || offset < 0 {
		return ListResult{}, ErrInvalidListArg
	}
	if limit == 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	query = strings.TrimSpace(query)
	if len(query) > MaxQueryLength {
		query = query[:MaxQueryLength]
	}
	if s.lister == nil {
		return ListResult{}, errors.New("user lister not configured")
	}

	users, total, err := s.lister.List(ctx, domain.ListFilter{Query: query, Limit: limit, Offset: offset})
	if err != nil {
		return ListResult{}, err
	}
	withRoles, err := s.attachRoles(ctx, users)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Users: withRoles, Total: total, Limit: limit, Offset: offset}, nil
}

// Get loads one user by numeric id or public uuid.
func (s *Service) Get(ctx context.Context, rawID string) (*UserWithRoles, error) {
	rawID = strings.TrimSpace(rawID)
	var (
		u   *domain.User
		err error
	)
	if n, convErr := strconv.ParseInt(rawID, 10, 64); convErr == nil {
		if n <= 0 {
			return nil, ErrInvalidUserID
		}
		u, err = s.repo.FindByID(domain.UserID(n))
	} else if _, parseErr := uuid.Parse(rawID); parseErr == nil {
		u, err = s.repo.FindByUUID(rawID)
	} else {
		return nil, ErrInvalidUserID
	}
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	out, err := s.attachRoles(ctx, []domain.User{*u})
	if err != nil {
		return nil, err
	}
	return &out[0], nil
}

func (s *Service) attachRoles(ctx context.Context, users []domain.User) ([]UserWithRoles, error) {
	out := make([]UserWithRoles, 0, len(users))
	var byUser map[int64][]string
	if s.roles != nil && len(users) > 0 {
		ids := make([]int64, 0, len(users))
		for _, u := range users {
			ids = append(ids, u.ID)
		}
		var err error
		byUser, err = s.roles.RoleNamesByUserIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
	}
	for _, u := range users {
		roles := byUser[u.ID]
		if roles == nil {
			roles = []string{}
		}
		out = append(out, UserWithRoles{User: u, Roles: roles})
	}
	return out, nil
}
