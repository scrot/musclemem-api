package move

import (
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

func NewMoveCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <command>",
		Short: "move resource position",
		Long:  `Move a resource its position in a list`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		NewMoveExerciseCmd(c),
	)
	return cmd
}
