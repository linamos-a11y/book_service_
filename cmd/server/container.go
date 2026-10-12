package main

import (
	"fmt"

	"github.com/fiverr/book_service_/internal/handler"
	"github.com/fiverr/book_service_/internal/middleware"
	"github.com/fiverr/book_service_/internal/repository/activitystore"
	esrepo "github.com/fiverr/book_service_/internal/repository/elasticsearch"
	"github.com/fiverr/book_service_/internal/router"
	"github.com/fiverr/book_service_/internal/service"
	"go.uber.org/dig"
)

// Compile-time checks that each concrete type satisfies the interface it is
// bound to in buildContainer. dig.As only fails at startup, so these catch
// a mismatch at build time instead.
var (
	_ service.BookRepository      = (*esrepo.BookRepository)(nil)
	_ service.ActivityRepository  = (*activitystore.ActivityRepository)(nil)
	_ handler.BookService         = (*service.BookService)(nil)
	_ handler.SearchService       = (*service.SearchService)(nil)
	_ handler.StoreService        = (*service.StoreService)(nil)
	_ handler.ActivityService     = (*service.ActivityService)(nil)
	_ middleware.ActivityRecorder = (*service.ActivityService)(nil)
)

// provider is a constructor plus the dig options (such as dig.As) that bind
// its concrete result to the interfaces its consumers declare.
type provider struct {
	constructor any
	opts        []dig.ProvideOption
}

func buildContainer() (*dig.Container, error) {
	container := dig.New()
	providers := []provider{
		{constructor: newConfigGetter},
		{constructor: provideLogger},
		// clients
		{constructor: provideESClient},
		{constructor: provideRedisClient},
		// repositories: concrete structs, bound to the interfaces the service layer declares
		{provideBookRepository, []dig.ProvideOption{dig.As(new(service.BookRepository))}},
		{activitystore.NewActivityRepository, []dig.ProvideOption{dig.As(new(service.ActivityRepository))}},
		// services: concrete structs, bound to the interfaces their consumers declare
		{service.NewBookService, []dig.ProvideOption{dig.As(new(handler.BookService))}},
		{service.NewSearchService, []dig.ProvideOption{dig.As(new(handler.SearchService))}},
		{service.NewStoreService, []dig.ProvideOption{dig.As(new(handler.StoreService))}},
		{service.NewActivityService, []dig.ProvideOption{dig.As(new(handler.ActivityService), new(middleware.ActivityRecorder))}},
		// handlers
		{constructor: handler.NewBooksHandler},
		{constructor: handler.NewSearchHandler},
		{constructor: handler.NewStoreHandler},
		{constructor: handler.NewActivityHandler},
		// middleware
		{constructor: middleware.Activity},
		// router: its Dependencies struct is assembled by dig (dig.In)
		{constructor: router.New},
	}
	for _, p := range providers {
		if err := container.Provide(p.constructor, p.opts...); err != nil {
			return nil, fmt.Errorf("providing %T: %w", p.constructor, err)
		}
	}
	return container, nil
}
