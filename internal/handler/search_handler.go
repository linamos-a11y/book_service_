package handler

import (
	"context"
	"net/http"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/fiverr/book_service_/internal/service"
	"github.com/fiverr/go_platform/core/logging"
	"github.com/gin-gonic/gin"
)

// SearchService is the search use case the handler depends on.
type SearchService interface {
	Search(ctx context.Context, input service.SearchInput) (books []model.Book, total int64, err error)
}

// SearchHandler handles GET /search.
type SearchHandler struct {
	svc    SearchService
	logger logging.Logger
}

// NewSearchHandler builds a SearchHandler backed by the given SearchService.
func NewSearchHandler(svc SearchService, logger logging.Logger) *SearchHandler {
	return &SearchHandler{svc: svc, logger: logger}
}

func (ctrl *SearchHandler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	books, total, err := ctrl.svc.Search(c.Request.Context(), service.SearchInput{
		Title:      req.Title,
		AuthorName: req.AuthorName,
		PriceRange: req.PriceRange,
	})
	if err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"total": total, "results": books})
}
