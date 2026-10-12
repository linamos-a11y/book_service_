package handler

import (
	"context"
	"net/http"

	"github.com/fiverr/book_service_/internal/model"
	"github.com/fiverr/book_service_/internal/service"
	"github.com/fiverr/go_platform/core/logging"
	"github.com/gin-gonic/gin"
)

// BookService is the book use case the handler depends on.
type BookService interface {
	Create(ctx context.Context, input service.CreateBookInput) (id string, err error)
	Update(ctx context.Context, id, title string) error
	Get(ctx context.Context, id string) (*model.Book, error)
	Delete(ctx context.Context, id string) error
}

// BooksHandler handles POST/PUT/GET/DELETE on /books.
type BooksHandler struct {
	svc    BookService
	logger logging.Logger
}

// NewBooksHandler builds a BooksHandler backed by the given BookService.
func NewBooksHandler(svc BookService, logger logging.Logger) *BooksHandler {
	return &BooksHandler{svc: svc, logger: logger}
}

func (ctrl *BooksHandler) Create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	id, err := ctrl.svc.Create(c.Request.Context(), service.CreateBookInput{
		Title:          req.Title,
		AuthorName:     req.AuthorName,
		Price:          *req.Price,
		EbookAvailable: req.EbookAvailable,
		PublishDate:    req.PublishDate,
	})
	if err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusCreated, gin.H{"id": id})
}

func (ctrl *BooksHandler) Update(c *gin.Context) {
	var req updateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	if err := ctrl.svc.Update(c.Request.Context(), req.ID, req.Title); err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": req.ID, "title": req.Title})
}

func (ctrl *BooksHandler) Get(c *gin.Context) {
	var req bookIDRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	book, err := ctrl.svc.Get(c.Request.Context(), req.ID)
	if err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusOK, book)
}

func (ctrl *BooksHandler) Delete(c *gin.Context) {
	var req bookIDRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, http.StatusBadRequest, bindErrorText(err))
		return
	}

	if err := ctrl.svc.Delete(c.Request.Context(), req.ID); err != nil {
		respondFromError(c, ctrl.logger, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": req.ID, "deleted": true})
}
