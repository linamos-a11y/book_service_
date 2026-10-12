package main

import (
	"fmt"
	"log"

	esconnector "github.com/fiverr/go_es_connector"
	"github.com/gin-gonic/gin"
)

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
