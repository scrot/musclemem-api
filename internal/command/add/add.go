package add

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

type AddOptions struct{}

func NewAddCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <command>",
		Short: "Add resource",
		Long: `Allows the creation of new user workouts
    and new workout exercises`,
		Example: heredoc.Doc(`
      $ mm add workout -f example/workouts.json
      $ mm add exercise 1 -f example/file.json
    `),
	}

	cmd.AddCommand(
		NewAddExerciseCmd(c),
		NewAddWorkoutCmd(c),
	)

	return cmd
}
