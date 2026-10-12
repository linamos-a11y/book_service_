package model

// Book is the domain representation of a book document, as stored in
// Elasticsearch and returned by the /books and /search endpoints.
type Book struct {
	ID             string  `json:"id,omitempty"`
	Title          string  `json:"title"`
	AuthorName     string  `json:"author_name"`
	Price          float64 `json:"price"`
	EbookAvailable bool    `json:"ebook_available"`
	PublishDate    string  `json:"publish_date"`
}
