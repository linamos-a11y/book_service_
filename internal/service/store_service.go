package service

import (
	"context"

	"github.com/fiverr/book_service_/internal/repository/elasticsearch"
)

// StoreStats holds the aggregated counters returned by GET /store.
type StoreStats = elasticsearch.StoreStats

// StoreService holds the business rules for GET /store and has no
// knowledge of HTTP.
type StoreService interface {
	Store(ctx context.Context) (StoreStats, error)
}

type storeService struct {
	repo elasticsearch.BookRepository
}

// NewStoreService builds a StoreService backed by the given BookRepository.
func NewStoreService(repo elasticsearch.BookRepository) StoreService {
	return &storeService{repo: repo}
}

func (s *storeService) Store(ctx context.Context) (StoreStats, error) {
	return s.repo.Store(ctx)
}
