package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/fiverr/book_service_/internal/service"
	"github.com/gin-gonic/gin"
)

func respondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, model.Response{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

func respondError(c *gin.Context, status int, text string) {
	c.JSON(status, model.Response{
		Success: false,
		Data:    nil,
		Error: &model.ErrorInfo{
			Text:       text,
			StatusCode: status,
		},
	})
}

// respondFromError maps a service-layer error to the matching HTTP status
// and unified error envelope.
func respondFromError(c *gin.Context, err error) {
	var validationErr *service.ValidationError
	switch {
	case errors.Is(err, service.ErrNotFound):
		respondError(c, http.StatusNotFound, "book not found")
	case errors.As(err, &validationErr):
		respondError(c, http.StatusBadRequest, validationErr.Text)
	default:
		log.Printf("unexpected error handling %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		respondError(c, http.StatusInternalServerError, "internal error")
	}
}

// requiredField pairs a request field's name (for the error message) with
// its value.
type requiredField struct {
	name  string
	value string
}

// requireFields checks that each field is non-empty, in order, responding
// with a 400 for the first missing one. Returns false if a field was
// missing (the caller should return immediately in that case).
func requireFields(c *gin.Context, fields ...requiredField) bool {
	for _, f := range fields {
		if f.value == "" {
			respondError(c, http.StatusBadRequest, f.name+" is required")
			return false
		}
	}
	return true
}
