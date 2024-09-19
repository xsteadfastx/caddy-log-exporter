package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr     string   `env:"ADDR,required"      envDefault:":2112"`
	LogFiles []string `env:"LOG_FILES,required"`
}

func Parse() (Config, error) {
	var cfg Config

	if err := env.ParseWithOptions(&cfg, env.Options{
		Prefix: "CADDY_LOG_EXPORTER_",
	}); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}
