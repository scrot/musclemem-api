package main

import (
	"log/slog"
	"os"
	"runtime"

	"github.com/caarlos0/env/v9"
	"github.com/lmittmann/tint"
	musclememapi "github.com/scrot/musclemem-api"
	"go.uber.org/automaxprocs/maxprocs"
)

var date string
var commit string
var version string
var maintainer string

func main() {
	opts := &tint.Options{
		Level: slog.LevelDebug,
	}
	log := slog.New(tint.NewHandler(os.Stdout, opts))

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

	// Setup database connections
	dbConfig := musclememapi.SqliteDatastoreConfig{
		DatabaseURL: "file://musclemem.db",
		Overwrite:   false,
		Logger:      nil,
	}

	store, err := musclememapi.NewSqliteDatastore(dbConfig)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	server := musclememapi.NewServer(cfg, log, store)
	server.Start()
}

// parseConfig parses global variables that can be set by LDFLAGS during build time
// environment variables overwrite build-time variables.
func parseConfig() (musclememapi.ServerConfig, error) {
	var cfg musclememapi.ServerConfig

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