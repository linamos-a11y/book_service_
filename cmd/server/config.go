package main

import (
	"fmt"

	esconnector "github.com/fiverr/go_es_connector"
	"github.com/fiverr/goenv"
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
