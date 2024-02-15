package edit

import (
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/command/edit/exercise"
	"github.com/scrot/musclemem-api/internal/command/edit/workout"
	"github.com/spf13/cobra"
)

type EditOptions struct{}

func NewEditCmd(c *cli.CLIConfig) *cobra.Command {
	// opts := EditOptions{}

	cmd := &cobra.Command{
		Use:   "edit <command>",
		Short: "Edit an user resource",
		Long:  `Edit an user resource, the user must be logged in`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		exercise.NewEditExerciseCmd(c),
		exercise.NewEditExerciseDownCmd(c),
		exercise.NewEditExerciseUpCmd(c),
		exercise.NewEditExerciseSwapCmd(c),
		workout.NewEditWorkoutCmd(c),
	)

	return cmd
}
