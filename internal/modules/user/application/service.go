package application

import "github.com/nextpresskit/backend/internal/modules/user/domain"

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}
