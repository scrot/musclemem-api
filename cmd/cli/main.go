package main

import (
	"context"
	"fmt"
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

// TODO: init configuration file (baseurl, configpath)
// TODO: list single exercise / workout using wi/ei?
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

	config.BaseURL = "http://localhost:8080"

	root := command.NewRootCmd(config)
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		return cli.ExitError
	}

	return cli.ExitOK
}
