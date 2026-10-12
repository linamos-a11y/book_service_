package router

import (
	"github.com/fiverr/book_service_/internal/handler"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// Dependencies bundles every handler and middleware the router needs to
// register its routes. The dig container fills it in from its providers.
type Dependencies struct {
	dig.In

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

	tracked := r.Group("", deps.ActivityMiddleware)
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
