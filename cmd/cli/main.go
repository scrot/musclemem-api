package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "musclemem",
		Usage: "CLI for musclemem-api",
		Authors: []*cli.Author{
			{
				Name:  "Roy de Wildt",
				Email: "roydewildt@gmail.com",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add new exercise",
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Path to json file",
						Action: func(ctx *cli.Context, p cli.Path) error {
							if _, err := os.Stat(p); err != nil {
								return errors.New("file doesn't exist")
							}
							return nil
						},
					},
				},
				Action: func(ctx *cli.Context) error {
					xs, err := os.ReadFile(ctx.Path("file"))
					if err != nil {
						return err
					}

					fmt.Println("added exercise: ", string(xs))
					return nil
				},
			},
		},
		Suggest: true,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
