package workout

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	model "github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
)

type EditWorkoutOptions struct {
	model.Workout
}

func NewEditWorkoutCmd(c *cli.CLIConfig) *cobra.Command {
	opts := EditWorkoutOptions{}

	cmd := &cobra.Command{
		Use:     "workout <index>",
		Aliases: []string{"wo"},
		Short:   "Edit a workout",
		Long: `Edit a existing workout belonging to a user,
    The workout must exist and the user must be logged-in.`,
		Example: heredoc.Doc(`
    $ mm edit workout 1 --name "Full-body workout"
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if opts.Name == "" {
				return cli.NewCLIError(errors.New("missing flags"))
			}

			var wi int
			_, err := fmt.Sscanf(args[0], "%d", wi)
			if err != nil {
				return cli.NewCLIError(err)
			}

			payload, err := json.Marshal(&opts.Workout)
			if err != nil {
				return cli.NewJSONError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d", c.User, wi)

			resp, error := cli.SendRequest("PATCH", c.BaseURL, endpoint, bytes.NewReader(payload))
			if error != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "change workout name")

	return cmd
}
