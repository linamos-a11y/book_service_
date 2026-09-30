package handler

import (
	"net/http"

	"github.com/fiverr/book_service_/internal/service"
	"github.com/gin-gonic/gin"
)

// BooksHandler handles POST/PUT/GET/DELETE on /books.
type BooksHandler struct {
	svc service.BookService
}

// NewBooksHandler builds a BooksHandler backed by the given BookService.
func NewBooksHandler(svc service.BookService) *BooksHandler {
	return &BooksHandler{svc: svc}
}

type createBookRequest struct {
	Title          string   `json:"title"`
	AuthorName     string   `json:"author_name"`
	Price          *float64 `json:"price"`
	EbookAvailable *bool    `json:"ebook_available"`
	PublishDate    string   `json:"publish_date"`
	Username       string   `json:"username"`
}

func (ctrl *BooksHandler) Create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if !requireFields(c,
		requiredField{"title", req.Title},
		requiredField{"author_name", req.AuthorName},
		requiredField{"publish_date", req.PublishDate},
		requiredField{"username", req.Username},
	) {
		return
	}
	if req.Price == nil {
		respondError(c, http.StatusBadRequest, "price is required")
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
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusCreated, gin.H{"id": id})
}

type updateBookRequest struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
}

func (ctrl *BooksHandler) Update(c *gin.Context) {
	var req updateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if !requireFields(c,
		requiredField{"id", req.ID},
		requiredField{"title", req.Title},
		requiredField{"username", req.Username},
	) {
		return
	}

	if err := ctrl.svc.Update(c.Request.Context(), req.ID, req.Title); err != nil {
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": req.ID, "title": req.Title})
}

func (ctrl *BooksHandler) Get(c *gin.Context) {
	id := c.Query("id")
	username := c.Query("username")

	if !requireFields(c,
		requiredField{"id", id},
		requiredField{"username", username},
	) {
		return
	}

	book, err := ctrl.svc.Get(c.Request.Context(), id)
	if err != nil {
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, book)
}

func (ctrl *BooksHandler) Delete(c *gin.Context) {
	id := c.Query("id")
	username := c.Query("username")

	if !requireFields(c,
		requiredField{"id", id},
		requiredField{"username", username},
	) {
		return
	}

	if err := ctrl.svc.Delete(c.Request.Context(), id); err != nil {
		respondFromError(c, err)
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"id": id, "deleted": true})
}
