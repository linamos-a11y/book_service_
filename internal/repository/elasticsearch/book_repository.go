package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/olivere/elastic/v7"
)

// ErrNotFound is returned when a book id does not exist in the index.
var ErrNotFound = errors.New("book not found")

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

// BookRepository is the only layer that talks to Elasticsearch on behalf of
// the book resource.
type BookRepository interface {
	Create(ctx context.Context, book model.Book) (id string, err error)
	Update(ctx context.Context, id, title string) error
	Get(ctx context.Context, id string) (*model.Book, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, params SearchParams) (books []model.Book, total int64, err error)
	Store(ctx context.Context) (StoreStats, error)
}

type bookRepository struct {
	client *elastic.Client
	index  string
}

// NewBookRepository builds a BookRepository backed by the given Elasticsearch
// client and index name.
func NewBookRepository(client *elastic.Client, index string) BookRepository {
	return &bookRepository{client: client, index: index}
}

func (r *bookRepository) Create(ctx context.Context, book model.Book) (string, error) {
	res, err := r.client.Index().
		Index(r.index).
		BodyJson(book).
		Do(ctx)
	if err != nil {
		return "", err
	}
	return res.Id, nil
}

func (r *bookRepository) Update(ctx context.Context, id, title string) error {
	_, err := r.client.Update().
		Index(r.index).
		Id(id).
		Doc(map[string]interface{}{"title": title}).
		Do(ctx)
	if elastic.IsNotFound(err) {
		return ErrNotFound
	}
	return err
}

func (r *bookRepository) Get(ctx context.Context, id string) (*model.Book, error) {
	res, err := r.client.Get().
		Index(r.index).
		Id(id).
		Do(ctx)
	if elastic.IsNotFound(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var book model.Book
	if err := unmarshalSource(res.Source, &book); err != nil {
		return nil, err
	}
	book.ID = res.Id
	return &book, nil
}

func (r *bookRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Delete().
		Index(r.index).
		Id(id).
		Do(ctx)
	if elastic.IsNotFound(err) {
		return ErrNotFound
	}
	return err
}

func (r *bookRepository) Search(ctx context.Context, params SearchParams) ([]model.Book, int64, error) {
	query := buildQuery(params)

	res, err := r.client.Search().
		Index(r.index).
		Query(query).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	books := make([]model.Book, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var book model.Book
		if err := unmarshalSource(hit.Source, &book); err != nil {
			return nil, 0, err
		}
		book.ID = hit.Id
		books = append(books, book)
	}

	return books, totalHits(res), nil
}

func buildQuery(params SearchParams) elastic.Query {
	query := elastic.NewBoolQuery()

	if params.Title != "" {
		query = query.Must(elastic.NewMatchQuery("title", params.Title))
	}
	if params.AuthorName != "" {
		query = query.Must(elastic.NewMatchQuery("author_name", params.AuthorName))
	}
	if params.PriceMin != nil || params.PriceMax != nil {
		rangeQuery := elastic.NewRangeQuery("price")
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

func (r *bookRepository) Store(ctx context.Context) (StoreStats, error) {
	const distinctAuthorsAgg = "distinct_authors"

	res, err := r.client.Search().
		Index(r.index).
		Size(0).
		Aggregation(distinctAuthorsAgg, elastic.NewCardinalityAggregation().Field("author_name.keyword")).
		Do(ctx)
	if err != nil {
		return StoreStats{}, err
	}

	cardinality, _ := res.Aggregations.Cardinality(distinctAuthorsAgg)
	var distinctAuthors int64
	if cardinality != nil && cardinality.Value != nil {
		distinctAuthors = int64(*cardinality.Value)
	}

	return StoreStats{
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
