package commands

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	removeExerciseCmd.Flags().IntVarP(&removeFromWorkout, "workout", "w", 0, "index of user workout (required)")
	removeExerciseCmd.MarkFlagRequired("workout")

	removeCmd.AddCommand(removeExerciseCmd, removeWorkoutCmd)

	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove a user resource",
	Long:    `Remove a user resource, the user must be logged-in`,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var removeFromWorkout int

var removeExerciseCmd = &cobra.Command{
	Use:     "exercise [exercise index]",
	Aliases: []string{"ex"},
	Short:   "Remove an exercise",
	Long:    `Remove an exercise from a workout of a user, the user must be logged-in`,
	Args:    cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = removeFromWorkout
		)

		exercise, err := strconv.Atoi(args[0])
		if err != nil {
			handleCLIError(err)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d", username, workout, exercise)
		handleResponse(doJSON(http.MethodDelete, baseurl, endpoint, nil))
	},
}

var removeWorkoutCmd = &cobra.Command{
	Use:     "workout [workout index]",
	Aliases: []string{"wo"},
	Short:   "Remove a workout",
	Long:    `Remove an workout of a user, the user must be logged-in`,
	Args:    cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		username := viper.GetString("user")

		workout, err := strconv.Atoi(args[0])
		if err != nil {
			handleCLIError(err)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d", username, workout)
		handleResponse(doJSON(http.MethodDelete, baseurl, endpoint, nil))
	},
}
