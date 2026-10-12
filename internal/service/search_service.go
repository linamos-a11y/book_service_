package service

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/fiverr/book_service_/internal/model"
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
type SearchService struct {
	repo BookRepository
}

// NewSearchService builds a SearchService backed by the given BookRepository.
func NewSearchService(repo BookRepository) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) ([]model.Book, int64, error) {
	if input.Title == "" && input.AuthorName == "" && input.PriceRange == "" {
		return nil, 0, &ValidationError{Text: "at least one of title, author_name, or price_range is required"}
	}

	params := SearchParams{
		Title:      input.Title,
		AuthorName: input.AuthorName,
	}

	if input.PriceRange != "" {
		minPrice, maxPrice, err := parsePriceRange(input.PriceRange)
		if err != nil {
			return nil, 0, err
		}
		params.PriceMin = &minPrice
		params.PriceMax = &maxPrice
	}

	return s.repo.Search(ctx, params)
}

// parsePriceRange parses a "<min>-<max>" string, e.g. "10-50", into its
// numeric bounds.
func parsePriceRange(priceRange string) (minPrice, maxPrice float64, err error) {
	priceRange = strings.TrimSpace(priceRange)
	if strings.HasPrefix(priceRange, "-") {
		return 0, 0, &ValidationError{Text: "price_range min must not be negative"}
	}

	minPart, maxPart, found := strings.Cut(priceRange, "-")
	if !found {
		return 0, 0, &ValidationError{Text: "price_range must be in the form \"<min>-<max>\", e.g. \"10-50\""}
	}

	minPrice, minErr := strconv.ParseFloat(strings.TrimSpace(minPart), 64)
	maxPrice, maxErr := strconv.ParseFloat(strings.TrimSpace(maxPart), 64)
	if minErr != nil || maxErr != nil || !isFinite(minPrice) || !isFinite(maxPrice) {
		return 0, 0, &ValidationError{Text: "price_range bounds must be numeric"}
	}
	if minPrice > maxPrice {
		return 0, 0, &ValidationError{Text: "price_range min must not be greater than max"}
	}

	return minPrice, maxPrice, nil
}

// isFinite reports whether f is a regular number (not NaN or +/-Inf).
func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
