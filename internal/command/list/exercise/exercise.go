package exercise

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	model "github.com/scrot/musclemem-api/internal/exercise"
	"github.com/spf13/cobra"
)

type ListExerciseOption struct{}

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
			wi, err := strconv.Atoi(args[0])
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises", c.User, wi)
			resp, err := cli.SendRequest(http.MethodGet, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			defer resp.Body.Close()
			dec := json.NewDecoder(resp.Body)

			var xs []model.Exercise
			if err := dec.Decode(&xs); err != nil {
				return cli.NewAPIError(err)
			}

			printTable(c, xs)

			return nil
		},
	}
	return cmd
}

func printTable(c *cli.CLIConfig, xs []model.Exercise) {
	t := cli.NewSimpleTable(c)
	t.SetHeader([]string{"#", "NAME", "WEIGHT", "REPS"})
	for i, x := range xs {
		t.Append([]string{
			fmt.Sprintf("%d", i+1),
			x.Name,
			fmt.Sprintf("%.1f", x.Weight),
			fmt.Sprintf("%d", x.Repetitions),
		})
	}
	t.Render()
}
