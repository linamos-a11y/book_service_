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

// Response is the unified top-level envelope returned by every endpoint.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *ErrorInfo  `json:"error"`
}

// ErrorInfo carries the human-readable message and mirrored HTTP status code
// for a failed request.
type ErrorInfo struct {
	Text       string `json:"text"`
	StatusCode int    `json:"status_code"`
}
