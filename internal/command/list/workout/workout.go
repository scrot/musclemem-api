package workout

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	model "github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
)

type ListWorkoutOption struct{}

func ListWorkoutCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workout",
		Aliases: []string{"wo"},
		Short:   "list workouts of user",
		Long:    `lists all workouts belonging the logged-in user`,
		Example: heredoc.Doc(`
      $ mm list workout
    `),
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			endpoint := fmt.Sprintf("/users/%s/workouts", c.User)
			resp, err := cli.SendRequest(http.MethodGet, c.BaseURL, endpoint, nil)
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			defer resp.Body.Close()
			dec := json.NewDecoder(resp.Body)

			var ws []model.Workout
			if err := dec.Decode(&ws); err != nil {
				return cli.NewJSONError(err)
			}

			printTable(c, ws)

			return nil
		},
	}

	return cmd
}

func printTable(c *cli.CLIConfig, ws []model.Workout) {
	t := cli.NewSimpleTable(c)
	t.SetHeader([]string{"INDEX", "NAME"})
	for _, w := range ws {
		t.Append([]string{strconv.Itoa(w.Index), w.Name})
	}
	t.Render()
}
