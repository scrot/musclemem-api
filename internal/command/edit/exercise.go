package edit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/spf13/cobra"
)

type EditExerciseOptions struct {
	exercise.Exercise
}

func NewEditExerciseCmd(c *cli.CLIConfig) *cobra.Command {
	opts := EditExerciseOptions{}

	cmd := &cobra.Command{
		Use:     "exercise <workout-index>/<exercise-index>",
		Aliases: []string{"ex"},
		Short:   "Edit an exercise (ref workout/exercise)",
		Long: `Edit a existing exercise belonging to a workout,
    to reference use workout-index/exercise-index. The workout
    and exercise must exist and the user must be logged-in.`,
		Example: heredoc.Doc(`
      $ mm edit exercise 1/1 --name "pull ups"
      $ mm edit exercise 1/2 --weight 40.5 --reps 15 
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if opts.Name == "" &&
				opts.Weight == 0 &&
				opts.Repetitions == 0 {
				return cli.NewCLIError(errors.New("missing flags"))
			}

			var wi, ei int
			_, err := fmt.Sscanf(args[0], "%d/%d", &wi, &ei)
			if err != nil {
				return cli.NewCLIError(err)
			}

			payload, err := json.Marshal(&opts.Exercise)
			if err != nil {
				return cli.NewJSONError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d", c.User, wi, ei)

			resp, error := cli.SendRequest("PATCH", c.BaseURL, endpoint, bytes.NewReader(payload))
			if error != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "change exercise name")
	cmd.Flags().Float64Var(&opts.Weight, "weight", 0, "change exercise weight")
	cmd.Flags().IntVar(&opts.Repetitions, "reps", 0, "change exercise repetitions")

	return cmd
}
