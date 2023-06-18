package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v8"
	"github.com/lmittmann/tint"
	"golang.org/x/exp/slog"
)

const version = "0.0.1"

type config struct {
	Port int    `env:"PORT" envDefault:"4000"`
	Env  string `env:"ENVIRONMENT" envDefault:"development"`
}

type api struct {
	config config
	logger *slog.Logger
}

func main() {
	log := slog.New(tint.NewHandler(os.Stdout, nil))

	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	api := &api{
		config: cfg,
		logger: log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", api.healthcheckHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info("starting api server", "port", cfg.Port, "env", cfg.Env)

	if err := srv.ListenAndServe(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
