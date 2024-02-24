package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/lmittmann/tint"
	"github.com/scrot/musclemem-api/internal"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/storage"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/scrot/musclemem-api/internal/workout"
	"go.uber.org/automaxprocs/maxprocs"
)

var version string

// TODO: return different ExitCodes (also for cancel)
func main() {
	if _, err := maxprocs.Set(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := run(ctx, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	env := getenv("ENVIRONMENT")

	// set logger with colour
	var opts tint.Options
	if env == "development" {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(tint.NewHandler(os.Stdout, &opts)).With("version", version)

	// configure database
	dbConfig := storage.DefaultSqliteConfig
	db, err := storage.NewSqliteDatastore(dbConfig)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	// register service stores
	us := user.NewSQLUserStore(db)
	ws := workout.NewSQLWorkoutStore(db)
	xs := exercise.NewSQLExerciseStore(db)

	// configure and start server
	cfg := internal.ServerConfig{
		ListenAddr: net.JoinHostPort(getenv("HOST"), getenv("PORT")),
	}

	server := internal.NewServer(cfg, l, us, ws, xs)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	// graceful shutdown on terminate signal
	go func() {
		kill := make(chan os.Signal, 1)
		signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
		<-kill
		l.Info("server received kill signal, shutting down")
		cancel()
	}()

	server.Start(ctx)

	return nil
}
