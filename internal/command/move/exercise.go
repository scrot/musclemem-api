package move

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

func NewMoveExerciseCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exercise <command>",
		Aliases: []string{"ex"},
		Short:   "Move exercise",
		Example: heredoc.Doc(`
      $ mm move exercise 1/2 up
      $ mm move exercise 1/1 down
      $ mm move exercise swap 1/1 1/2
    `),
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(
		NewMoveExerciseUpCmd(c),
		NewMoveExerciseDownCmd(c),
		NewMoveExerciseSwapCmd(c),
	)

	return cmd
}

func NewMoveExerciseDownCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down <workout-index>/<exercise-index>",
		Short: "Move exercise down",
		Long: `Move an exercise down in the list of workout exercises
    if the exercise is already the last exercise then nothing happens`,
		Example: heredoc.Doc(`
      $ mm move exercise down 1/2
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi, ei int
			_, err := fmt.Sscanf(args[0], "%d/%d", &wi, &ei)
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/down", c.User, wi, ei)
			resp, err := cli.SendRequest(http.MethodPut, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp)
			}

			return nil
		},
	}

	return cmd
}

func NewMoveExerciseUpCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up <workout-index>/<exercise-index>",
		Short: "Move exercise up",
		Long: `Move an exercise up in the list of workout exercises
    if the exercise is already the first exercise then nothing happens`,
		Example: heredoc.Doc(`
      $ mm move exercise up 1/2
    `),
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi, ei int
			_, err := fmt.Sscanf(args[0], "%d/%d", &wi, &ei)
			if err != nil {
				return cli.NewCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/up", c.User, wi, ei)
			resp, err := cli.SendRequest(http.MethodPut, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp)
			}

			return nil
		},
	}

	return cmd
}

func NewMoveExerciseSwapCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap <workout-index>/<exercise-index> <workout-index>/<exercise-index>",
		Short: "swap two exercises",
		Long: `swap the exercise provided by the first argument 
    with the exercise from the second argument. Only exercises
    within the same workout can be swapped`,
		Example: heredoc.Doc(`
      $ mm move exercise swap 1/2 1/3
    `),
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			var wi1, ei1 int
			if _, err := fmt.Sscanf(args[0], "%d/%d", &wi1, &ei1); err != nil {
				return cli.NewCLIError(err)
			}

			var wi2, ei2 int
			if _, err := fmt.Sscanf(args[1], "%d/%d", &wi2, &ei2); err != nil {
				return cli.NewCLIError(err)
			}

			if wi1 != wi2 {
				return cli.NewCLIError(errors.New("excercises not of same workout"))
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/swap", c.User, wi1, ei1)
			body := strings.NewReader(fmt.Sprintf("{%q: %d}", "with", ei2))

			resp, err := cli.SendRequest(http.MethodPost, c.BaseURL, endpoint, body)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp)
			}

			return nil
		},
	}

	return cmd
}
