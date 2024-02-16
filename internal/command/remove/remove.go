package remove

import (
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

type RemoveOptions struct{}

func NewRemoveCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <command>",
		Aliases: []string{"rm"},
		Short:   "Remove a user resource",
		Long:    `Remove a user resource, the user must be logged-in`,
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		NewRemoveExerciseCmd(c),
		NewRemoveWorkoutCmd(c),
	)

	return cmd
}
