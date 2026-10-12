package handler

import (
	"context"
	"net/http"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/fiverr/go_platform/core/logging"
	"github.com/gin-gonic/gin"
)

// StoreService is the store statistics use case the handler depends on.
type StoreService interface {
	Store(ctx context.Context) (service.StoreStats, error)
}

// StoreHandler handles GET /store.
type StoreHandler struct {
	svc    StoreService
	logger logging.Logger
}

// NewStoreHandler builds a StoreHandler backed by the given StoreService.
func NewStoreHandler(svc StoreService, logger logging.Logger) *StoreHandler {
	return &StoreHandler{svc: svc, logger: logger}
}

func (ctrl *StoreHandler) Store(c *gin.Context) {
	var req usernameRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	stats, err := ctrl.svc.Store(c.Request.Context())
	if err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"total_books":      stats.TotalBooks,
		"distinct_authors": stats.DistinctAuthors,
	})
}
