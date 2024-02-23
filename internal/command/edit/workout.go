package edit

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
)

type EditWorkoutOptions struct {
	workout.Workout
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
			ref, err := workout.ParseRef(c.User + "/" + args[0])
			if err != nil {
				return cli.NewCLIError(err)
			}

			if _, _, err := c.Workouts.Update(context.TODO(), ref, opts.Workout); err != nil {
				return cli.NewAPIError(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "change workout name")

	return cmd
}
