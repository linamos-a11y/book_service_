package service

import (
	"context"

	"github.com/fiverr/book_service_/internal/model"
)

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
type BookService struct {
	repo BookRepository
}

// NewBookService builds a BookService backed by the given BookRepository.
func NewBookService(repo BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) Create(ctx context.Context, input CreateBookInput) (string, error) {
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

func (s *BookService) Update(ctx context.Context, id, title string) error {
	return s.repo.Update(ctx, id, title)
}

func (s *BookService) Get(ctx context.Context, id string) (*model.Book, error) {
	return s.repo.Get(ctx, id)
}

func (s *BookService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
