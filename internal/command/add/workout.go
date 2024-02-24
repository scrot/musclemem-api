package add

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/api"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
)

type AddWorkoutOptions struct {
	FilePath string
}

// NewAddWorkoutCmd is the cli command for adding exercises to a user workout
// it should only be used in combination with the NewAddCmd
func NewAddWorkoutCmd(c *cli.CLIConfig) *cobra.Command {
	opts := AddWorkoutOptions{}

	cmd := &cobra.Command{
		Use:     "workout",
		Aliases: []string{"wo"},
		Short:   "Add one or more workouts",
		Long:    `Add a new workout to the list of user workouts `,
		Example: heredoc.Doc(`
      # add single workout from json file
      $ mm add workout -f path/to/workout.json

      # add multiple workouts from json file
      $ mm add workout -f path/to/workouts.json
    `),
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			file, err := os.ReadFile(opts.FilePath)
			if err != nil {
				return cli.NewCLIError(err)
			}

			dec := json.NewDecoder(bytes.NewReader(file))

			var ws []workout.Workout
			switch api.JSONType(file) {
			case api.TypeJSONObject:
				var w workout.Workout
				if err := dec.Decode(&w); err != nil {
					return cli.NewCLIError(err)
				}
				ws = append(ws, w)
			case api.TypeJSONArray:
				if err := dec.Decode(&ws); err != nil {
					return cli.NewCLIError(err)
				}
			default:
				err := errors.New("invalid json type")
				return cli.NewCLIError(err)
			}

			for _, w := range ws {
				w.Owner = c.User
				if _, _, err := c.Workouts.Add(context.TODO(), w); err != nil {
					return cli.NewAPIError(err)
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.FilePath, "file", "f", "", "path to json file (required)")
	cmd.MarkPersistentFlagRequired("file")

	return cmd
}
