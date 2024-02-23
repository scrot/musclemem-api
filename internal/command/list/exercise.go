package list

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
)

type ListExerciseOptions struct{}

func ListExerciseCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exercise <workout-index>",
		Aliases: []string{"ex"},
		Short:   "list exercises of a workout",
		Long: `lists all exercises belonging to a workout index
    only exercises can be listed that belongs to a workout
    of the logged-in user`,
		Example: heredoc.Doc(`
    $ mm list exercise 1
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ref, err := workout.ParseRef(c.User + "/" + args[0])
			if err != nil {
				return cli.NewCLIError(err)
			}

			slog.Info("parsed exercise", "ref", ref)

			xs, _, err := c.Exercises.List(context.TODO(), ref)
			if err != nil {
				return cli.NewAPIError(err)
			}

			t := cli.NewSimpleTable(c)
			t.SetHeader([]string{"#", "NAME", "WEIGHT", "REPS"})
			for _, x := range xs {
				t.Append([]string{
					strconv.Itoa(x.Index),
					x.Name,
					fmt.Sprintf("%.1f", x.Weight),
					fmt.Sprintf("%d", x.Repetitions),
				})
			}
			t.Render()

			return nil
		},
	}
	return cmd
}
