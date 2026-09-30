package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/fiverr/book_service_/internal/repository/elasticsearch"
)

// SearchInput carries the optional filters accepted by GET /search.
// At least one of Title, AuthorName, or PriceRange must be non-empty.
type SearchInput struct {
	Title      string
	AuthorName string
	PriceRange string
}

// SearchService holds the business rules for GET /search and has no
// knowledge of HTTP.
type SearchService interface {
	Search(ctx context.Context, input SearchInput) (books []model.Book, total int64, err error)
}

type searchService struct {
	repo elasticsearch.BookRepository
}

// NewSearchService builds a SearchService backed by the given BookRepository.
func NewSearchService(repo elasticsearch.BookRepository) SearchService {
	return &searchService{repo: repo}
}

func (s *searchService) Search(ctx context.Context, input SearchInput) ([]model.Book, int64, error) {
	if input.Title == "" && input.AuthorName == "" && input.PriceRange == "" {
		return nil, 0, &ValidationError{Text: "at least one of title, author_name, or price_range is required"}
	}

	params := elasticsearch.SearchParams{
		Title:      input.Title,
		AuthorName: input.AuthorName,
	}

	if input.PriceRange != "" {
		min, max, err := parsePriceRange(input.PriceRange)
		if err != nil {
			return nil, 0, err
		}
		params.PriceMin = &min
		params.PriceMax = &max
	}

	return s.repo.Search(ctx, params)
}

// parsePriceRange parses a "<min>-<max>" string, e.g. "10-50", into its
// numeric bounds.
func parsePriceRange(priceRange string) (min, max float64, err error) {
	priceRange = strings.TrimSpace(priceRange)
	if strings.HasPrefix(priceRange, "-") {
		return 0, 0, &ValidationError{Text: "price_range min must be greater than 0"}
	}

	minPart, maxPart, found := strings.Cut(priceRange, "-")
	if !found {
		return 0, 0, &ValidationError{Text: "price_range must be in the form \"<min>-<max>\", e.g. \"10-50\""}
	}

	min, minErr := strconv.ParseFloat(strings.TrimSpace(minPart), 64)
	max, maxErr := strconv.ParseFloat(strings.TrimSpace(maxPart), 64)
	if minErr != nil || maxErr != nil {
		return 0, 0, &ValidationError{Text: "price_range bounds must be numeric"}
	}
	if min <= 0 {
		return 0, 0, &ValidationError{Text: "price_range min must be greater than 0"}
	}

	return min, max, nil
}
