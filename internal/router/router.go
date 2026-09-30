package router

import (
	"github.com/fiverr/book_service_/internal/handler"
	"github.com/gin-gonic/gin"
)

// Dependencies bundles every handler and middleware the router needs to
// register its routes.
type Dependencies struct {
	Books    *handler.BooksHandler
	Search   *handler.SearchHandler
	Store    *handler.StoreHandler
	Activity *handler.ActivityHandler
	// ActivityMiddleware runs on every route except /activity itself.
	ActivityMiddleware gin.HandlerFunc
}

// New builds the gin.Engine and registers all routes.
func New(deps Dependencies) *gin.Engine {
	r := gin.Default()

	r.GET("/activity", deps.Activity.Get)

	tracked := r.Group("/")
	tracked.Use(deps.ActivityMiddleware)
	{
		books := tracked.Group("/books")
		{
			books.POST("", deps.Books.Create)
			books.PUT("", deps.Books.Update)
			books.GET("", deps.Books.Get)
			books.DELETE("", deps.Books.Delete)
		}

		tracked.GET("/search", deps.Search.Search)
		tracked.GET("/store", deps.Store.Store)
	}

	return r
}
