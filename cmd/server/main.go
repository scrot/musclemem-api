package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/lmittmann/tint"
	"github.com/scrot/musclemem-api/internal"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/storage"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
	"go.uber.org/automaxprocs/maxprocs"
)

var version string

type Env struct {
	URL         string `env:"URL" envDefault:":8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

// TODO: return different ExitCodes (also for cancel)
func main() {
	if _, err := maxprocs.Set(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, _ io.Writer, _ []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// load environment variables
	var environment Env
	if err := env.Parse(&environment); err != nil {
		os.Exit(1)
	}

	// set logger with colour
	var opts tint.Options
	if environment.Environment == "development" {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(tint.NewHandler(os.Stdout, &opts)).With("version", version)

	// adhere container quota for cores if set
	procs := runtime.GOMAXPROCS(0)

	dbConfig := storage.DefaultSqliteConfig
	db, err := storage.NewSqliteDatastore(dbConfig)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	us := user.NewSQLUserStore(db)
	ws := workout.NewSQLWorkoutStore(db)
	xs := exercise.NewSQLExerciseStore(db)

	server := internal.NewServer(l, us, ws, xs)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:         environment.URL,
		Handler:      server,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// run server in its own goroutine
	go func() {
		l.Info("server started listening", "addr", environment.URL, "cores", procs, "env", environment.Environment)
		if err := httpServer.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				l.Info("server stopped listening to new requests")
			default:
				l.Error(fmt.Sprintf("unexpected error while listening: %s", err))
			}
		}
	}()

	// graceful shutdown on terminate signal
	go func() {
		kill := make(chan os.Signal, 1)
		signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
		<-kill
		cancel()
	}()

	// graceful shutdown pattern adopted from Mat Ryer
	// shuts down the server if context get's canceled
	done := make(chan struct{}, 1)

	go func() {
		<-ctx.Done()
		// new context required since ctx is already canceled
		shutdownCtx := context.Background()
		shutdownCtx, cancel = context.WithTimeout(shutdownCtx, 3*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			l.Error(fmt.Sprintf("error shutting down http server: %s", err))
			os.Exit(1)
		}
		l.Info("server gracefully shutdown, till next time!")
		done <- struct{}{}
	}()

	<-done

	return nil
}
