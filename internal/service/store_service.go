package service

import "context"

// StoreService holds the business rules for GET /store and has no
// knowledge of HTTP.
type StoreService struct {
	repo BookRepository
}

// NewStoreService builds a StoreService backed by the given BookRepository.
func NewStoreService(repo BookRepository) *StoreService {
	return &StoreService{repo: repo}
}

func (s *StoreService) Store(ctx context.Context) (StoreStats, error) {
	return s.repo.Store(ctx)
}
