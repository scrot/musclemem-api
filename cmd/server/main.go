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

var (
	version string
	env     string
)

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

	// set logger
	var lh slog.Handler
	if env == "development" {
		lh = tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug})
	} else {
		lh = gcHandler()
	}
	l := slog.New(lh).With("version", version)

	// configure database
	var dbConfig storage.DatastoreConfig
	if env == "development" {
		dbConfig = storage.DefaultSqliteConfig
	} else {
		dbConfig = storage.DatastoreConfig{
			DatabaseURL:   getenv("DATABASE_DSN"),
			MigrationPath: "migrations",
		}
	}

	db, err := storage.NewSqlDatastore(dbConfig)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	// register service stores
	us := user.NewSQLUserStore(db)
	ws := workout.NewSQLWorkoutStore(db)
	xs := exercise.NewSQLExerciseStore(db)

	// configure and start server
	var cfg internal.ServerConfig
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if env == "development" {
		cfg = internal.ServerConfig{
			ListenAddr: ":" + port,
		}
	} else {
		cfg = internal.ServerConfig{
			ListenAddr: net.JoinHostPort("0.0.0.0", port),
		}
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

func gcHandler() slog.Handler {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.MessageKey:
				a.Key = "message"
			case slog.SourceKey:
				a.Key = "logging.googleapis.com/sourceLocation"
			case slog.LevelKey:
				a.Key = "severity"
				level := a.Value.Any().(slog.Level)
				if level == slog.Level(12) {
					a.Value = slog.StringValue("CRITICAL")
				}
			}
			return a
		},
	})
	return h
}
