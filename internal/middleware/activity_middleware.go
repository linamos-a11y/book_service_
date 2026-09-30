package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/gin-gonic/gin"
)

// Activity returns a middleware that records, per user, the HTTP method and
// route of every request it wraps. It is registered on every route except
// /activity itself, so that looking up one's activity does not count as an
// action.
func Activity(svc service.ActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := extractUsername(c)

		c.Next()

		if username == "" {
			return
		}

		_ = svc.RecordAction(c.Request.Context(), username, c.Request.Method, c.FullPath())
	}
}

// extractUsername reads "username" from the query string (GET/DELETE) or,
// for POST/PUT, peeks at the JSON body without consuming it, so the
// downstream handler can still bind the full body itself.
func extractUsername(c *gin.Context) string {
	if username := c.Query("username"); username != "" {
		return username
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return payload.Username
}
