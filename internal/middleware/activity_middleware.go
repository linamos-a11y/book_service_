package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/fiverr/go_platform/core/logging"
	"github.com/gin-gonic/gin"
)

// ActivityRecorder is the part of the activity use case this middleware
// depends on.
type ActivityRecorder interface {
	RecordAction(ctx context.Context, username, method, path string) error
}

// Activity returns a middleware that records, per user, the HTTP method and
// route of every request it wraps. It is registered on every route except
// /activity itself, so that looking up one's activity does not count as an
// action.
func Activity(svc ActivityRecorder, logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := extractUsername(c)

		c.Next()

		if username == "" {
			return
		}

		// The response has usually been written by now, so the request context
		// may already be cancelled. Keep its values but not its cancellation.
		ctx := context.WithoutCancel(c.Request.Context())
		if err := svc.RecordAction(ctx, username, c.Request.Method, c.FullPath()); err != nil {
			logger.WithContext(ctx).Error("failed to record user activity",
				logging.String("method", c.Request.Method),
				logging.String("route", c.FullPath()),
				logging.Error(err))
		}
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
