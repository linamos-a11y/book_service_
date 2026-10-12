package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/fiverr/book_service_/internal/service"
	"github.com/olivere/elastic/v7"
)

// Field names in the books index mapping.
const (
	fieldTitle         = "title"
	fieldAuthorName    = "author_name"
	fieldAuthorKeyword = "author_name.keyword"
	fieldPrice         = "price"
)

// BookRepository is the only layer that talks to Elasticsearch on behalf of
// the book resource. It implements service.BookRepository.
type BookRepository struct {
	client *elastic.Client
	index  string
}

// NewBookRepository builds a BookRepository backed by the given Elasticsearch
// client and index name.
func NewBookRepository(client *elastic.Client, index string) *BookRepository {
	return &BookRepository{client: client, index: index}
}

func (r *BookRepository) Create(ctx context.Context, book model.Book) (string, error) {
	res, err := r.client.Index().
		Index(r.index).
		BodyJson(book).
		Do(ctx)
	if err != nil {
		return "", fmt.Errorf("index book: %w", err)
	}
	return res.Id, nil
}

func (r *BookRepository) Update(ctx context.Context, id, title string) error {
	_, err := r.client.Update().
		Index(r.index).
		Id(id).
		Doc(map[string]interface{}{fieldTitle: title}).
		Do(ctx)
	if elastic.IsNotFound(err) {
		return service.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}
	return nil
}

func (r *BookRepository) Get(ctx context.Context, id string) (*model.Book, error) {
	res, err := r.client.Get().
		Index(r.index).
		Id(id).
		Do(ctx)
	if elastic.IsNotFound(err) {
		return nil, service.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	var book model.Book
	if err := unmarshalSource(res.Source, &book); err != nil {
		return nil, fmt.Errorf("decode book: %w", err)
	}
	book.ID = res.Id
	return &book, nil
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Delete().
		Index(r.index).
		Id(id).
		Do(ctx)
	if elastic.IsNotFound(err) {
		return service.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	return nil
}

func (r *BookRepository) Search(ctx context.Context, params service.SearchParams) ([]model.Book, int64, error) {
	query := buildQuery(params)

	res, err := r.client.Search().
		Index(r.index).
		Query(query).
		Do(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("search books: %w", err)
	}

	books := make([]model.Book, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var book model.Book
		if err := unmarshalSource(hit.Source, &book); err != nil {
			return nil, 0, fmt.Errorf("decode search hit: %w", err)
		}
		book.ID = hit.Id
		books = append(books, book)
	}

	return books, totalHits(res), nil
}

func buildQuery(params service.SearchParams) elastic.Query {
	query := elastic.NewBoolQuery()

	if params.Title != "" {
		query = query.Must(elastic.NewMatchQuery(fieldTitle, params.Title))
	}
	if params.AuthorName != "" {
		query = query.Must(elastic.NewMatchQuery(fieldAuthorName, params.AuthorName))
	}
	if params.PriceMin != nil || params.PriceMax != nil {
		rangeQuery := elastic.NewRangeQuery(fieldPrice)
		if params.PriceMin != nil {
			rangeQuery = rangeQuery.Gte(*params.PriceMin)
		}
		if params.PriceMax != nil {
			rangeQuery = rangeQuery.Lte(*params.PriceMax)
		}
		query = query.Filter(rangeQuery)
	}

	return query
}

func (r *BookRepository) Store(ctx context.Context) (service.StoreStats, error) {
	const distinctAuthorsAgg = "distinct_authors"

	res, err := r.client.Search().
		Index(r.index).
		Size(0).
		Aggregation(distinctAuthorsAgg, elastic.NewCardinalityAggregation().Field(fieldAuthorKeyword)).
		Do(ctx)
	if err != nil {
		return service.StoreStats{}, fmt.Errorf("aggregate store stats: %w", err)
	}

	cardinality, _ := res.Aggregations.Cardinality(distinctAuthorsAgg)
	var distinctAuthors int64
	if cardinality != nil && cardinality.Value != nil {
		distinctAuthors = int64(*cardinality.Value)
	}

	return service.StoreStats{
		TotalBooks:      totalHits(res),
		DistinctAuthors: distinctAuthors,
	}, nil
}

func totalHits(res *elastic.SearchResult) int64 {
	if res.Hits == nil || res.Hits.TotalHits == nil {
		return 0
	}
	return res.Hits.TotalHits.Value
}

func unmarshalSource(source json.RawMessage, book *model.Book) error {
	return json.Unmarshal(source, book)
}
