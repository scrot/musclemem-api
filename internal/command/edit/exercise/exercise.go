package exercise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	model "github.com/scrot/musclemem-api/internal/exercise"
	"github.com/spf13/cobra"
)

type EditExerciseOptions struct {
	model.Exercise
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
			_, err := fmt.Sscanf(args[0], "%d/%d", wi, ei)
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
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "change exercise name")
	cmd.Flags().Float64Var(&opts.Weight, "weight", 0, "change exercise weight")
	cmd.Flags().IntVar(&opts.Repetitions, "reps", 0.0, "change exercise repetitons")

	return cmd
}

func NewEditExerciseDownCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down <workout-index>/<exercise-index>",
		Short: "Move exercise down",
		Long: `Move an exercise down in the list of workout exercises
    if the exercise is already the last exercise then nothing happens`,
		Example: heredoc.Doc(`
    $ mm edit exercise down 1/2
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi, ei int
			_, err := fmt.Sscanf(args[0], "%d/%d", wi, ei)
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/down", c.User, wi, ei)
			resp, err := cli.SendRequest(http.MethodPut, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	return cmd
}

func NewEditExerciseUpCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up <workout-index>/<exercise-index>",
		Short: "Move exercise up",
		Long: `Move an exercise up in the list of workout exercises
    if the exercise is already the first exercise then nothing happens`,
		Example: heredoc.Doc(`
    $ mm edit exercise up 1/2
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi, ei int
			_, err := fmt.Sscanf(args[0], "%d/%d", wi, ei)
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/up", c.User, wi, ei)
			resp, err := cli.SendRequest(http.MethodPut, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	return cmd
}

func NewEditExerciseSwapCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap <workout-index>/<exercise-index> <workout-index>/<exercise-index>",
		Short: "swap two exercises",
		Long: `swap the exercise provided by the first argument 
    with the exercise from the second argument. Only exercises
    within the same workout can be swapped`,
		Example: heredoc.Doc(`
    $ mm edit exercise swap 1/2 1/3
    `),
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi1, ei1 int
			if _, err := fmt.Sscanf(args[0], "%d/%d", wi1, ei1); err != nil {
				return cli.NewCLIError(err)
			}

			var wi2, ei2 int
			if _, err := fmt.Sscanf(args[1], "%d/%d", wi2, ei2); err != nil {
				return cli.NewCLIError(err)
			}

			if wi1 != wi2 {
				return cli.NewCLIError(errors.New("excercises not in the same workout"))
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/swap", c.BaseURL, wi1, ei1)
			body := strings.NewReader(fmt.Sprintf("{%q: %d}", "with", ei2))

			resp, err := cli.SendRequest(http.MethodPost, c.BaseURL, endpoint, body)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			return nil
		},
	}

	return cmd
}
