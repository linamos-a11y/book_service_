package service

import (
	"context"

	"github.com/fiverr/book_service_/internal/model"
)

// SearchParams carries the optional filters accepted by GET /search.
// At least one field is expected to be set by the caller (enforced by the
// service layer); the repository applies whichever are non-nil.
type SearchParams struct {
	Title      string
	AuthorName string
	PriceMin   *float64
	PriceMax   *float64
}

// StoreStats holds the aggregated counters returned by GET /store.
type StoreStats struct {
	TotalBooks      int64
	DistinctAuthors int64
}

// BookRepository is what the book, search, and store services need from the
// storage layer for the book resource.
type BookRepository interface {
	Create(ctx context.Context, book model.Book) (id string, err error)
	Update(ctx context.Context, id, title string) error
	Get(ctx context.Context, id string) (*model.Book, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, params SearchParams) (books []model.Book, total int64, err error)
	Store(ctx context.Context) (StoreStats, error)
}

// ActivityRepository is what the activity service needs from the storage
// layer for the per-user activity log.
type ActivityRepository interface {
	RecordAction(ctx context.Context, username, action string) error
	GetActions(ctx context.Context, username string) ([]string, error)
}
