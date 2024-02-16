package main

import (
	"context"
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

func main() {
	// TODO: create good main / run seperation with args
	// return proper ExitCode
	run()
}

func run() {
	// load environment variables
	var vs Env
	if err := env.Parse(&vs); err != nil {
		os.Exit(1)
	}

	// set logger with colour
	var opts tint.Options
	if vs.Environment == "development" {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(tint.NewHandler(os.Stdout, &opts)).With("version", version)

	// adhere container quota for cores if set
	if _, err := maxprocs.Set(); err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
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
		Addr:         vs.URL,
		Handler:      server,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	ctx := context.Background()

	// TODO: make pretty with graceful shutdown etc
	l.Info("starting api server", "addr", vs.URL, "cores", procs, "env", vs.Environment)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				l.Info("api server stopped listening to new requests")
			default:
				l.Error(err.Error())
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
		l.Error(err.Error())
		os.Exit(1)
	}

	l.Info("server is gracefully shutting down", "timeout", d)
	ctx, cancel := context.WithTimeout(ctx, d)
	if err := httpServer.Shutdown(ctx); err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}
	cancel()
}
