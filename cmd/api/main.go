package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/lmittmann/tint"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/exp/slog"
)

type config struct {
	Version         string `env:"VERSION" envDefault:"0.0.1"`
	Port            int    `env:"PORT" envDefault:"4000"`
	Env             string `env:"ENVIRONMENT" envDefault:"development"`
	ShutdownTimeout string `env:"SHUTDOWN_TIMEOUT" envDefault:"3s"`
}

type api struct {
	config config
	logger *slog.Logger
}

func main() {
	ctx := context.Background()

	log := slog.New(tint.NewHandler(os.Stdout, nil))

	// set GOMAXPROCS to adhere container quota if set
	if _, err := maxprocs.Set(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	g := runtime.GOMAXPROCS(0)

	// parse environment variables
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	d, err := time.ParseDuration(cfg.ShutdownTimeout)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	api := &api{
		config: cfg,
		logger: log,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      api.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Info("starting api server", "port", cfg.Port, "env", cfg.Env, "cpus", g)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("api server stopped listening to new requests")
			default:
				log.Error(err.Error())
			}
		}
	}()

	// block till termination signal is received
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	// shutdown server or kill server after timeout expires
	log.Info("server is gracefully shutting down", "timeout", d)
	ctx, cancel := context.WithTimeout(ctx, d)
	if err := srv.Shutdown(ctx); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	cancel()
}
