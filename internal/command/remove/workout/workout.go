package workout

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

type RemoveWorkoutOptions struct{}

func NewRemoveWorkoutCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workout <workout-index>",
		Aliases: []string{"wo"},
		Short:   "Remove a workout",
		Long:    `Remove a workout, the user must be logged-in`,
		Example: heredoc.Doc(`
    $ mm remove workout 1
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi int
			_, err := fmt.Sscanf(args[0], "%d", wi)
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d", c.User, wi)
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
