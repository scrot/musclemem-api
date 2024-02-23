package remove

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/exercise"
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
			ref, err := exercise.ParseRef(c.User + "/" + args[0])
			if err != nil {
				return cli.NewCLIError(err)
			}

			if _, err := c.Exercises.Delete(context.TODO(), ref); err != nil {
				return cli.NewAPIError(err)
			}

			return nil
		},
	}

	return cmd
}
