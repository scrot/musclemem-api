package exercise

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

type AddExerciseOptions struct {
	FilePath string
}

// NewAddExerciseCmd is the cli command for adding exercises to a user workout
// it should only be used in combination with the NewAddCmd
func NewAddExerciseCmd(c *cli.CLIConfig) *cobra.Command {
	opts := AddExerciseOptions{}

	cmd := &cobra.Command{
		Use:     "exercise <workout-index>",
		Aliases: []string{"ex"},
		Short:   "Add one or more exercises",
		Long:    `Add a new exercise to provided workout index for the signed in user`,
		Args:    cobra.ExactArgs(1),
		Example: heredoc.Doc(`
      # add single exercise from json file
      $ mm add exercise 1 -f path/to/exercise.json

      # add multiple exercises from json file
      $ mm add exercise 1 -f path/to/exercises.json
      `),
		RunE: func(_ *cobra.Command, args []string) error {
			file, err := os.Open(opts.FilePath)
			if err != nil {
				return cli.NewCLIError(err)
			}
			defer file.Close()

			wi, err := strconv.Atoi(args[0])
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises", c.User, wi)
			resp, err := cli.SendRequest(http.MethodPost, c.BaseURL, endpoint, file)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.FilePath, "file", "f", "", "path to json file (required)")
	cmd.MarkPersistentFlagRequired("file")

	return cmd
}
