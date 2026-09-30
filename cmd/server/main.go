package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fiverr/book_service_/internal/handler"
	"github.com/fiverr/book_service_/internal/middleware"
	esrepo "github.com/fiverr/book_service_/internal/repository/elasticsearch"
	activityrepo "github.com/fiverr/book_service_/internal/repository/redis"
	"github.com/fiverr/book_service_/internal/router"
	"github.com/fiverr/book_service_/internal/service"
	esconnector "github.com/fiverr/go_es_connector"
	"github.com/fiverr/goenv"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/dig"
)

const (
	readTimeout     = 10 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
)

// configGetter reads values from the templated config file selected by
// GO_CONFIG and GO_ENV, the same way go_searcher does.
type configGetter struct{ cfg *goenv.Goenv }

func (c configGetter) Get(key, def string) string {
	if value := c.cfg.Get(key, ""); value != "" {
		return value
	}
	return def
}

func newConfigGetter() esconnector.EnvGetter {
	return configGetter{cfg: goenv.DefaultTemplateGoenv()}
}

func requireConfig(g esconnector.EnvGetter, key string) (string, error) {
	value := g.Get(key, "")
	if value == "" {
		return "", fmt.Errorf("missing required config key %q", key)
	}
	return value, nil
}

func provideESClient(g esconnector.EnvGetter) (*esconnector.ElasticClient, error) {
	c, err := esconnector.NewConnectionManager(g).Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}
	return c, nil
}

func provideRedisClient(g esconnector.EnvGetter) (*goredis.Client, error) {
	host, err := requireConfig(g, "redis.default.host")
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}
	port, err := requireConfig(g, "redis.default.port")
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}
	return goredis.NewClient(&goredis.Options{Addr: net.JoinHostPort(host, port)}), nil
}

// provideBookRepository adapts the connector's client (which carries both the
// elastic client and the index name) to NewBookRepository's two parameters.
func provideBookRepository(c *esconnector.ElasticClient) esrepo.BookRepository {
	return esrepo.NewBookRepository(c.Client, c.Index)
}

// provideRouterDependencies bundles the handlers and middleware for router.New.
func provideRouterDependencies(
	books *handler.BooksHandler,
	search *handler.SearchHandler,
	store *handler.StoreHandler,
	activity *handler.ActivityHandler,
	activityMiddleware gin.HandlerFunc,
) router.Dependencies {
	return router.Dependencies{
		Books:              books,
		Search:             search,
		Store:              store,
		Activity:           activity,
		ActivityMiddleware: activityMiddleware,
	}
}

func buildContainer() (*dig.Container, error) {
	container := dig.New()
	constructors := []any{
		newConfigGetter,
		// clients
		provideESClient, provideRedisClient,
		// repositories
		provideBookRepository, activityrepo.NewActivityRepository,
		// services
		service.NewBookService, service.NewSearchService,
		service.NewStoreService, service.NewActivityService,
		// handlers
		handler.NewBooksHandler, handler.NewSearchHandler,
		handler.NewStoreHandler, handler.NewActivityHandler,
		// middleware
		middleware.Activity,
		// router
		provideRouterDependencies, router.New,
	}
	for _, constructor := range constructors {
		if err := container.Provide(constructor); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func serve(engine *gin.Engine, addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(stop)

	select {
	case err := <-serveErr:
		return fmt.Errorf("failed to listen and serve: %w", err)
	case <-stop:
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shut down server: %w", err)
	}
	return nil
}

func main() {
	container, err := buildContainer()
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = container.Invoke(func(engine *gin.Engine, g esconnector.EnvGetter) error {
		addr, err := requireConfig(g, "server.addr")
		if err != nil {
			return fmt.Errorf("failed to resolve server address: %w", err)
		}
		return serve(engine, addr)
	})
	if err != nil {
		log.Fatalf("server failed: %s", err)
	}
}
