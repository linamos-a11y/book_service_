package handler

import (
	"net/http"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/gin-gonic/gin"
)

// StoreHandler handles GET /store.
type StoreHandler struct {
	svc service.StoreService
}

// NewStoreHandler builds a StoreHandler backed by the given StoreService.
func NewStoreHandler(svc service.StoreService) *StoreHandler {
	return &StoreHandler{svc: svc}
}

func (ctrl *StoreHandler) Store(c *gin.Context) {
	username := c.Query("username")
	if !requireFields(c, requiredField{"username", username}) {
		return
	}

	stats, err := ctrl.svc.Store(c.Request.Context())
	if err != nil {
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{
		"total_books":      stats.TotalBooks,
		"distinct_authors": stats.DistinctAuthors,
	})
}
