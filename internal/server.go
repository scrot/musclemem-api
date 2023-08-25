package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/exp/slog"
)

type ServerConfig struct {
	Version         string `env:"VERSION"`
	Date            string `env:"DATE"`
	Maintainer      string `env:"MAINTAINER"`
	Commit          string `env:"COMMIT"`
	Threads         int    `env:"THREADS"`
	Port            int    `env:"PORT" envDefault:"4000"`
	Environment     string `env:"ENVIRONMENT" envDefault:"development"`
	ShutdownTimeout string `env:"SHUTDOWN_TIMEOUT" envDefault:"3s"`
}

type Server struct {
	ServerConfig
	logger *slog.Logger
}

func NewServer(cfg ServerConfig, logger *slog.Logger) *Server {
	return &Server{cfg, logger}
}

func (s *Server) Start() {
	ctx := context.Background()

	router := http.NewServeMux()
	router.HandleFunc("/health", s.HandleHealth)

	d, err := time.ParseDuration(s.ShutdownTimeout)
	if err != nil {
		s.logger.Error(err.Error())
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	s.logger.Info("starting api server", "port", s.Port, "env", s.Environment, "cpus", s.Threads)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				s.logger.Info("api server stopped listening to new requests")
			default:
				s.logger.Error(err.Error())
			}
		}
	}()

	// block till termination signal is received
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	// shutdown server or kill server after timeout expires
	s.logger.Info("server is gracefully shutting down", "timeout", d)
	ctx, cancel := context.WithTimeout(ctx, d)
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error(err.Error())
		os.Exit(1)
	}
	cancel()
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "status: available\n")
	fmt.Fprintf(w, "environment: %s\n", s.Environment)
	fmt.Fprintf(w, "version: %s\n", s.Version)
	fmt.Fprintf(w, "commit: %s\n", s.Commit)
	fmt.Fprintf(w, "date: %s\n", s.Date)
	fmt.Fprintf(w, "maintainer: %s\n", s.Maintainer)
}
