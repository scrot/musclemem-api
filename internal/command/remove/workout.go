package remove

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/workout"
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
			ref, err := workout.ParseRef(c.User + "/" + args[0])
			if err != nil {
				return cli.NewCLIError(err)
			}

			if _, _, err := c.Workouts.Delete(context.TODO(), ref); err != nil {
				return cli.NewAPIError(err)
			}

			return nil
		},
	}

	return cmd
}
