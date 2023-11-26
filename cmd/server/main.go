package main

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v9"
	"github.com/lmittmann/tint"
	"github.com/scrot/musclemem-api/internal"
	"go.uber.org/automaxprocs/maxprocs"
)

type Env struct {
	URL         string `env:"URL" envDefault:"127.0.0.1:8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

func main() {
	var vs Env
	if err := env.Parse(&vs); err != nil {
		os.Exit(1)
	}

	// set logger
	var opts tint.Options
	if vs.Environment == "development" {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(tint.NewHandler(os.Stdout, &opts))

	// set GOMAXPROCS to adhere container quota if set
	if _, err := maxprocs.Set(); err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	server, err := internal.NewServer(l)
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	server.Start()
}
