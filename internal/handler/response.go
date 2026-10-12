package handler

import (
	"errors"
	"net/http"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/fiverr/go_platform/core/logging"
	"github.com/gin-gonic/gin"
)

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

func respondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

func respondError(c *gin.Context, status int, text string) {
	c.JSON(status, Response{
		Success: false,
		Data:    nil,
		Error: &ErrorInfo{
			Text:       text,
			StatusCode: status,
		},
	})
}

// respondFromError maps a service-layer error to the matching HTTP status
// and unified error envelope. Unexpected errors are logged through logger.
func respondFromError(c *gin.Context, logger logging.Logger, err error) {
	var validationErr *service.ValidationError
	switch {
	case errors.Is(err, service.ErrNotFound):
		respondError(c, http.StatusNotFound, "book not found")
	case errors.As(err, &validationErr):
		respondError(c, http.StatusBadRequest, validationErr.Text)
	default:
		logger.WithContext(c.Request.Context()).Error("unexpected error handling request",
			logging.String("method", c.Request.Method),
			logging.String("route", c.FullPath()),
			logging.Error(err))
		respondError(c, http.StatusInternalServerError, "internal error")
	}
}
