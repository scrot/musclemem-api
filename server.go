package musclememapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type (
	Server struct {
		config ServerConfig
		logger *slog.Logger
		store  Datastore
	}

	ServerConfig struct {
		Version         string `env:"VERSION"`
		Date            string `env:"DATE"`
		Maintainer      string `env:"MAINTAINER"`
		Commit          string `env:"COMMIT"`
		Threads         int    `env:"THREADS"`
		URL             string `env:"URL" envDefault:"127.0.0.1:8080"`
		Environment     string `env:"ENVIRONMENT" envDefault:"development"`
		ShutdownTimeout string `env:"SHUTDOWN_TIMEOUT" envDefault:"3s"`
	}

	Datastore interface {
		ExerciseRetreiver
		ExerciseStorer
	}
)

func NewServer(cfg ServerConfig, logger *slog.Logger, store Datastore) *Server {
	return &Server{
		cfg,
		logger,
		store,
	}
}

func (s *Server) Routes() http.Handler {
	router := chi.NewRouter()

	router.MethodFunc(http.MethodGet, "/health", s.HandleHealth)
	router.MethodFunc(http.MethodGet, "/exercises/{exerciseID}", s.HandleSingleExercise)

	return router
}

func (s *Server) Start() {
	ctx := context.Background()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s", s.config.URL),
		Handler:      s.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	s.logger.Info("starting api server", "addr", srv.Addr, "env", s.config.Environment, "cpus", s.config.Threads)
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
	d, err := time.ParseDuration(s.config.ShutdownTimeout)
	if err != nil {
		s.logger.Error(err.Error())
		os.Exit(1)
	}

	s.logger.Info("server is gracefully shutting down", "timeout", d)
	ctx, cancel := context.WithTimeout(ctx, d)
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Error(err.Error())
		os.Exit(1)
	}
	cancel()
}
