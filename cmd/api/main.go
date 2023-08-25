package main

import (
	"os"
	"runtime"

	"github.com/caarlos0/env/v9"
	"github.com/lmittmann/tint"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/exp/slog"

	"github.com/scrot/musclemem-api/internal"
)

var date string
var commit string
var version string
var maintainer string

func main() {
	log := slog.New(tint.NewHandler(os.Stdout, nil))

	// set GOMAXPROCS to adhere container quota if set
	if _, err := maxprocs.Set(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Parse environment variables and build flags
	cfg, err := parseConfig()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	server := internal.NewServer(cfg, log)
	server.Start()
}

// parseConfig parses global variables that can be set by LDFLAGS during build time
// environment variables overwrite build-time variables.
func parseConfig() (internal.ServerConfig, error) {
	var cfg internal.ServerConfig

	g := runtime.GOMAXPROCS(0)
	cfg.Threads = g

	if version != "" {
		cfg.Version = version
	} else {
		cfg.Version = "v0.0.0"
	}

	if maintainer != "" {
		cfg.Maintainer = maintainer
	}

	if date != "" {
		cfg.Date = date
	}

	if commit != "" {
		cfg.Commit = commit
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
