package service

import (
	"context"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/fiverr/book_service_/internal/repository/elasticsearch"
)

// ValidationError signals a client-input problem that should surface as a
// 400 response (as opposed to a not-found or infrastructure error).
type ValidationError struct {
	Text string
}

func (e *ValidationError) Error() string {
	return e.Text
}

// ErrNotFound is re-exported so callers only need to import the service
// package to distinguish not-found from other errors.
var ErrNotFound = elasticsearch.ErrNotFound

// CreateBookInput carries the fields accepted by POST /books.
// EbookAvailable is a pointer so the service can tell "omitted" (nil, to be
// defaulted to false) apart from an explicit false.
type CreateBookInput struct {
	Title          string
	AuthorName     string
	Price          float64
	EbookAvailable *bool
	PublishDate    string
}

// BookService holds the business rules for creating, updating, reading, and
// deleting a single book, and has no knowledge of HTTP.
type BookService interface {
	Create(ctx context.Context, input CreateBookInput) (id string, err error)
	Update(ctx context.Context, id, title string) error
	Get(ctx context.Context, id string) (*model.Book, error)
	Delete(ctx context.Context, id string) error
}

type bookService struct {
	repo elasticsearch.BookRepository
}

// NewBookService builds a BookService backed by the given BookRepository.
func NewBookService(repo elasticsearch.BookRepository) BookService {
	return &bookService{repo: repo}
}

func (s *bookService) Create(ctx context.Context, input CreateBookInput) (string, error) {
	ebookAvailable := false
	if input.EbookAvailable != nil {
		ebookAvailable = *input.EbookAvailable
	}

	book := model.Book{
		Title:          input.Title,
		AuthorName:     input.AuthorName,
		Price:          input.Price,
		EbookAvailable: ebookAvailable,
		PublishDate:    input.PublishDate,
	}

	return s.repo.Create(ctx, book)
}

func (s *bookService) Update(ctx context.Context, id, title string) error {
	return s.repo.Update(ctx, id, title)
}

func (s *bookService) Get(ctx context.Context, id string) (*model.Book, error) {
	return s.repo.Get(ctx, id)
}

func (s *bookService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
