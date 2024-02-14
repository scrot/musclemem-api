package main

import (
	"context"
	"os"

	"github.com/scrot/musclemem-api/internal/cli"
	command "github.com/scrot/musclemem-api/internal/command/root"
)

var (
	name    = "musclemem"
	version = "0.0.1"
	author  = "Roy de Wildt"
	date    = ""
)

// TODO: list single exercise / workout using wi/ei?
// TODO: refs to exercise like {user}/{workout}/{exercise}
// TODO: make it possible to change config path
// TODO: create a test for each command
// TODO: see if build variables are loaded correctly
// TODO: implement cancel context correctly
func main() {
	code := mainRun()
	os.Exit(int(code))
}

func mainRun() cli.ExitCode {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := cli.NewCLIConfig(name, version, author, date)
	if err != nil {
		return cli.ExitError
	}

	root := command.NewRootCmd(config)
	if err := root.ExecuteContext(ctx); err != nil {
		return cli.ExitError
	}

	return cli.ExitOK
}
