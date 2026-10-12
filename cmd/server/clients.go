package main

import (
	"fmt"
	"net"

	esrepo "github.com/fiverr/book_service_/internal/repository/elasticsearch"
	esconnector "github.com/fiverr/go_es_connector"
	"github.com/redis/go-redis/v9"
)

func provideESClient(g esconnector.EnvGetter) (*esconnector.ElasticClient, error) {
	c, err := esconnector.NewConnectionManager(g).Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}
	return c, nil
}

func provideRedisClient(g esconnector.EnvGetter) (*redis.Client, error) {
	host, err := requireConfig(g, "redis.default.host")
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}
	port, err := requireConfig(g, "redis.default.port")
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}
	return redis.NewClient(&redis.Options{Addr: net.JoinHostPort(host, port)}), nil
}

// provideBookRepository adapts the connector's client (which carries both the
// elastic client and the index name) to NewBookRepository's two parameters.
func provideBookRepository(c *esconnector.ElasticClient) *esrepo.BookRepository {
	return esrepo.NewBookRepository(c.Client, c.Index)
}
