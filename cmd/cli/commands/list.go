package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/scrot/musclemem-api/internal/exercise"
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
	Use:     "exercises [workout index]",
	Aliases: []string{"ex"},
	Args:    cobra.ExactArgs(1),
	Short:   "list all workout exercises",
	Long: `pretty print all exercises belonging to a workout,
  only workouts that belong to that user can be listed`,
	Run: func(_ *cobra.Command, args []string) {
		var (
			user    = viper.GetString("user")
			workout = args[0]
		)

		endpoint := fmt.Sprintf("/users/%s/workouts/%s/exercises", user, workout)
		resp, err := doJSON(http.MethodGet, baseurl, endpoint, nil)
		if err != nil {
			fmt.Printf("api error: %s\n", err)
			os.Exit(1)
		}

		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)

		var xs []exercise.Exercise
		if err := dec.Decode(&xs); err != nil {
			fmt.Printf("json error: %s\n", err)
			os.Exit(1)
		}

		t := newTable()
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

		endpoint := fmt.Sprintf("/users/%s/workouts", u)
		resp, err := doJSON(http.MethodGet, baseurl, endpoint, nil)
		if err != nil {
			fmt.Printf("api error: %s\n", err)
			os.Exit(1)
		}

		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)

		var ws []workout.Workout
		if err := dec.Decode(&ws); err != nil {
			fmt.Printf("json error: %s\n", err)
			os.Exit(1)
		}

		t := newTable()
		t.SetHeader([]string{"INDEX", "NAME"})
		for _, w := range ws {
			t.Append([]string{strconv.Itoa(w.Index), w.Name})
		}

		t.Render()
	},
}
