package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	listCmd.AddCommand(
		listExercisesCmd,
		listWorkoutsCmd,
	)

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list items belonging to the logged-in user",
	Long: `print or export a list of exercises from a specific
  workout or workouts belonging to the logged-in user`,
	Run: func(_ *cobra.Command, _ []string) {
	},
}

var listExercisesCmd = &cobra.Command{
	Use:     "exercises",
	Aliases: []string{"ex"},
	Args:    cobra.ExactArgs(1),
	Short:   "list all workout exercises",
	Long: `pretty print all exercises belonging to a workout,
  only workouts that belong to that user can be listed`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var listWorkoutsCmd = &cobra.Command{
	Use:     "workouts",
	Aliases: []string{"wo"},
	Args:    cobra.NoArgs,
	Short:   "list all user workouts",
	Long:    `pretty print all workouts belonging to the logged-in user`,
	Run: func(_ *cobra.Command, _ []string) {
		u := viper.GetString("user")

		payload, err := getJSON(baseurl, fmt.Sprintf("/users/%s/workouts", u))
		if err != nil {
			fmt.Printf("api error: %s\n", err)
			os.Exit(1)
		}

		var ws []workout.Workout
		if err := json.Unmarshal(payload, &ws); err != nil {
			fmt.Printf("json error: %s\n", err)
			os.Exit(1)
		}

		t := tablewriter.NewWriter(os.Stdout)
		t.SetHeader([]string{"Name"})
		t.SetBorder(true)

		for _, w := range ws {
			t.Append([]string{w.Name})
		}

		t.Render()
	},
}
