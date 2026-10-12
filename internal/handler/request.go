package handler

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	// Report validation errors by their JSON/query name (author_name), not by
	// the Go field name (AuthorName).
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(fieldName)
	}
}

func fieldName(f reflect.StructField) string {
	for _, tag := range []string{"json", "form"} {
		if name := strings.Split(f.Tag.Get(tag), ",")[0]; name != "" && name != "-" {
			return name
		}
	}
	return f.Name
}

// bindErrorText turns a gin binding error into a short, client-facing message.
func bindErrorText(err error) string {
	var fieldErrs validator.ValidationErrors
	if !errors.As(err, &fieldErrs) || len(fieldErrs) == 0 {
		return "invalid request body"
	}

	fe := fieldErrs[0]
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "gte":
		return fe.Field() + " must be " + fe.Param() + " or greater"
	case "datetime":
		return fe.Field() + " must be a date in the form YYYY-MM-DD"
	default:
		return fe.Field() + " is invalid"
	}
}

// createBookRequest is the JSON body of POST /books.
type createBookRequest struct {
	Title          string   `json:"title" binding:"required"`
	AuthorName     string   `json:"author_name" binding:"required"`
	Price          *float64 `json:"price" binding:"required,gte=0"`
	EbookAvailable *bool    `json:"ebook_available"`
	PublishDate    string   `json:"publish_date" binding:"required,datetime=2006-01-02"`
	Username       string   `json:"username" binding:"required"`
}

// updateBookRequest is the JSON body of PUT /books.
type updateBookRequest struct {
	ID       string `json:"id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Username string `json:"username" binding:"required"`
}

// bookIDRequest is the query string of GET and DELETE /books.
type bookIDRequest struct {
	ID       string `form:"id" binding:"required"`
	Username string `form:"username" binding:"required"`
}

// searchRequest is the query string of GET /search. At least one of the
// optional filters is required, which the service enforces.
type searchRequest struct {
	Username   string `form:"username" binding:"required"`
	Title      string `form:"title"`
	AuthorName string `form:"author_name"`
	PriceRange string `form:"price_range"`
}

// usernameRequest is the query string of GET /store and GET /activity.
type usernameRequest struct {
	Username string `form:"username" binding:"required"`
}
