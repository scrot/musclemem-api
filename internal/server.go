package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/scrot/musclemem-api/internal/exercise"
)

type (
	Server struct {
		logger     *slog.Logger
		exercises  exercise.Service
		listenAddr string
	}
)

func NewServer(l *slog.Logger, listenAddr string) (*Server, error) {
	dbConfig := DefaultSqliteConfig
	db, err := NewSqliteDatastore(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("NewServer: new SQL database: %w", err)
	}

	xs := exercise.NewSQLExercises(db)
	exerciseSvc := exercise.NewService(l, xs)

	s := Server{
		logger:     l,
		exercises:  *exerciseSvc,
		listenAddr: listenAddr,
	}

	return &s, nil
}

func (s *Server) Routes() http.Handler {
	router := chi.NewRouter()

	router.MethodFunc(http.MethodGet, "/exercises/{exerciseID}", s.exercises.HandleSingleExercise)
	router.MethodFunc(http.MethodPost, "/exercises/", s.exercises.HandleNewExercise)

	return router
}

func (s *Server) Start() {
	ctx := context.Background()

	srv := &http.Server{
		Addr:         s.listenAddr,
		Handler:      s.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	procs := runtime.GOMAXPROCS(0)
	s.logger.Info("starting api server", "addr", srv.Addr, "cpus", procs)

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
	d, err := time.ParseDuration("3s")
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
