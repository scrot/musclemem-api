package exercise

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

type RemoveExerciseOptions struct{}

func NewRemoveExerciseCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exercise <workout-index>/<exercise-index>",
		Aliases: []string{"ex"},
		Short:   "Remove an exercise",
		Long:    `Remove an exercise from a workout of a user, the user must be logged-in`,
		Example: heredoc.Doc(`
    $ mm remove exercise 1/2
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi, ei int
			_, err := fmt.Sscanf(args[0], "%d/%d", wi, ei)
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d", c.User, wi, ei)
			resp, err := cli.SendRequest(http.MethodDelete, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	return cmd
}
