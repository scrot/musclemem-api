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

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/storage"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
)

type (
	Server struct {
		logger     *slog.Logger
		users      user.Service
		workouts   workout.Service
		exercises  exercise.Service
		listenAddr string
	}
)

func NewServer(l *slog.Logger, listenAddr string) (*Server, error) {
	dbConfig := storage.DefaultSqliteConfig
	db, err := storage.NewSqliteDatastore(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("NewServer: new SQL database: %w", err)
	}

	us := user.NewSQLUsers(db)
	userSvc := user.NewService(l, us)

	ws := workout.NewSQLWorkouts(db)
	workoutSvc := workout.NewService(l, ws)

	xs := exercise.NewSQLExercises(db)
	exerciseSvc := exercise.NewService(l, xs)

	s := Server{
		logger:     l,
		users:      *userSvc,
		workouts:   *workoutSvc,
		exercises:  *exerciseSvc,
		listenAddr: listenAddr,
	}

	return &s, nil
}

func (s *Server) Routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("POST /users", s.users.HandleNewUser)
	router.HandleFunc("GET /users/{username}/workouts", s.workouts.HandleWorkouts)
	router.HandleFunc("POST /users/{username}/workouts", s.workouts.HandleNewWorkout)
	router.HandleFunc("DELETE /users/{username}/workouts/{workout}", s.workouts.HandleDeleteWorkout)
	router.HandleFunc("PATCH /users/{username}/workouts/{workout}", s.workouts.HandleChangeWorkout)
	router.HandleFunc("GET /users/{username}/workouts/{workout}/exercises", s.exercises.HandleExercises)
	router.HandleFunc("POST /users/{username}/workouts/{workout}/exercises", s.exercises.HandleNewExercise)
	router.HandleFunc("GET /users/{username}/workouts/{workout}/exercises/{exercise}", s.exercises.HandleSingleExercise)
	router.HandleFunc("PATCH /users/{username}/workouts/{workout}/exercises/{exercise}", s.exercises.HandleChangeExercise)
	router.HandleFunc("PUT /users/{username}/workouts/{workout}/exercises/{exercise}/up", s.exercises.HandleMoveUpExercise)
	router.HandleFunc("PUT /users/{username}/workouts/{workout}/exercises/{exercise}/down", s.exercises.HandleMoveDownExercise)
	router.HandleFunc("POST /users/{username}/workouts/{workout}/exercises/{exercise}/swap", s.exercises.HandleSwapExercises)

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
