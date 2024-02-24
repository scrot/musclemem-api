package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

type Server struct {
	ServerConfig
	logger *slog.Logger
	mux    http.Handler
}

type ServerConfig struct {
	ListenAddr string
}

func NewServer(
	config ServerConfig,
	logger *slog.Logger,
	users user.UserStore,
	workouts workout.WorkoutStore,
	exercises exercise.ExerciseStore,
) *Server {
	mux := http.NewServeMux()
	RegisterEndpoints(mux, logger, users, workouts, exercises)
	return &Server{ServerConfig: config, logger: logger, mux: mux}
}

func (s *Server) Start(ctx context.Context) {
	httpServer := &http.Server{
		Addr:         s.ListenAddr,
		Handler:      s.mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// run server in its own goroutine
	go func() {
		procs := runtime.GOMAXPROCS(0)
		s.logger.Info("server started listening", "addr", s.ListenAddr, "cores", procs)
		if err := httpServer.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				s.logger.Info("server stopped listening to new requests")
			default:
				s.logger.Error(fmt.Sprintf("unexpected error while listening: %s", err))
			}
		}
	}()

	// graceful shutdown pattern adopted from Mat Ryer
	// shuts down the server if context get's canceled
	done := make(chan struct{}, 1)

	go func() {
		<-ctx.Done()
		// new context required since ctx is already canceled
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 3*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error(fmt.Sprintf("error shutting down http server: %s", err))
			os.Exit(1)
		}
		s.logger.Info("server gracefully shutdown, till next time!")
		done <- struct{}{}
	}()

	<-done
}
