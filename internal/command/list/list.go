package list

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/command/list/exercise"
	"github.com/scrot/musclemem-api/internal/command/list/workout"
	"github.com/spf13/cobra"
)

type ListOptions struct{}

func NewListCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <command>",
		Aliases: []string{"ls"},
		Short:   "list user resources",
		Long:    `lists all resources of the logged-in user`,
		Example: heredoc.Doc(`
      $ mm list workout
      $ mm list exercise 1
    `),
	}

	cmd.AddCommand(
		exercise.ListExerciseCmd(c),
		workout.ListWorkoutCmd(c),
	)

	return cmd
}
