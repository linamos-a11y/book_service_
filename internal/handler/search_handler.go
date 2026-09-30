package handler

import (
	"net/http"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/gin-gonic/gin"
)

// SearchHandler handles GET /search.
type SearchHandler struct {
	svc service.SearchService
}

// NewSearchHandler builds a SearchHandler backed by the given SearchService.
func NewSearchHandler(svc service.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

func (ctrl *SearchHandler) Search(c *gin.Context) {
	username := c.Query("username")
	if !requireFields(c, requiredField{"username", username}) {
		return
	}

	books, total, err := ctrl.svc.Search(c.Request.Context(), service.SearchInput{
		Title:      c.Query("title"),
		AuthorName: c.Query("author_name"),
		PriceRange: c.Query("price_range"),
	})
	if err != nil {
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"total": total, "results": books})
}
