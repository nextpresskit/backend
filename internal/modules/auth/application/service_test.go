package application

import (
	"context"
	"errors"
	"testing"

	userDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
)

type userRepoStub struct {
	byEmail map[string]*userDomain.User
	created *userDomain.User
}

func (s *userRepoStub) FindByID(_ userDomain.UserID) (*userDomain.User, error) { return nil, nil }
func (s *userRepoStub) FindByEmail(email string) (*userDomain.User, error) {
	return s.byEmail[email], nil
}
func (s *userRepoStub) Create(user *userDomain.User) error {
	s.created = user
	if s.byEmail == nil {
		s.byEmail = map[string]*userDomain.User{}
	}
	s.byEmail[user.Email] = user
	return nil
}
func (s *userRepoStub) Update(_ *userDomain.User) error { return nil }
func (s *userRepoStub) Delete(_ userDomain.UserID) error { return nil }

type tokenStub struct {
	accessErr        error
	refreshErr       error
	parseRefreshErr  error
	parseRefreshUser string
}

func (s tokenStub) GenerateAccessToken(userID string) (string, error) {
	if s.accessErr != nil {
		return "", s.accessErr
	}
	return "acc-" + userID, nil
}

func (s tokenStub) GenerateRefreshToken(userID string) (string, error) {
	if s.refreshErr != nil {
		return "", s.refreshErr
	}
	return "ref-" + userID, nil
}

func (s tokenStub) ParseAccessToken(_ string) (string, error) { return "", nil }

func (s tokenStub) ParseRefreshToken(_ string) (string, error) {
	if s.parseRefreshErr != nil {
		return "", s.parseRefreshErr
	}
	return s.parseRefreshUser, nil
}

type hasherStub struct {
	checkErr error
}

func (h hasherStub) HashPassword(plain string) (string, error) { return "hash:" + plain, nil }

func (h hasherStub) CheckPassword(_, _ string) error { return h.checkErr }

func TestRegister_CreatesUser(t *testing.T) {
	repo := &userRepoStub{byEmail: map[string]*userDomain.User{}}
	svc := NewService(repo, tokenStub{}, hasherStub{})

	u, err := svc.Register(context.Background(), "A", "B", "A@Example.COM", "password123")
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if u == nil || repo.created == nil {
		t.Fatal("expected user to be created")
	}
	if u.Email != "a@example.com" {
		t.Fatalf("expected normalized email, got %q", u.Email)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	repo := &userRepoStub{
		byEmail: map[string]*userDomain.User{
			"taken@example.com": {ID: "u1", Email: "taken@example.com"},
		},
	}
	svc := NewService(repo, tokenStub{}, hasherStub{})

	_, err := svc.Register(context.Background(), "A", "B", "taken@example.com", "password123")
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := &userRepoStub{
		byEmail: map[string]*userDomain.User{
			"user@example.com": {ID: "u1", Email: "user@example.com", Password: "hash"},
		},
	}
	svc := NewService(repo, tokenStub{}, hasherStub{})

	access, refresh, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("expected tokens, got access=%q refresh=%q", access, refresh)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	repo := &userRepoStub{byEmail: map[string]*userDomain.User{}}
	svc := NewService(repo, tokenStub{}, hasherStub{})

	_, _, err := svc.Login(context.Background(), "missing@example.com", "password123")
	if !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin for missing user, got %v", err)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc := NewService(&userRepoStub{}, tokenStub{parseRefreshErr: errors.New("bad token")}, hasherStub{})

	_, _, err := svc.Refresh(context.Background(), "bad")
	if !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin for invalid refresh, got %v", err)
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	svc := NewService(&userRepoStub{}, tokenStub{parseRefreshErr: errors.New("bad token")}, hasherStub{})

	err := svc.Logout(context.Background(), "bad")
	if !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin for invalid refresh on logout, got %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	svc := NewService(&userRepoStub{}, tokenStub{parseRefreshUser: "u1"}, hasherStub{})

	if err := svc.Logout(context.Background(), "good"); err != nil {
		t.Fatalf("unexpected logout error: %v", err)
	}
}
