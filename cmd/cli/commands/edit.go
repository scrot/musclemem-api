package commands

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	editExerciseCmd.PersistentFlags().IntVarP(&editWorkout, "workout", "w", 0, "index of user workout (required)")
	editExerciseCmd.MarkPersistentFlagRequired("workout")

	editExerciseCmd.AddCommand(exerciseDownCmd, exerciseUpCmd, exerciseSwapCmd)

	editCmd.AddCommand(editExerciseCmd, editWorkoutCmd)

	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an user resource",
	Long:  `Edit an user resource, the user must be logged in`,
	Args:  cobra.NoArgs,
}

var editWorkout int

var editExerciseCmd = &cobra.Command{
	Use:     "exercise",
	Aliases: []string{"ex"},
	Short:   "Edit an exercise",
	Long:    `Edit an exercise of a user workout, the user must be logged in`,
	Args:    cobra.NoArgs,
}

var editWorkoutCmd = &cobra.Command{
	Use:     "workout",
	Aliases: []string{"wo"},
	Short:   "Edit an workout",
	Long:    `Edit an user workout, the user must be logged in`,
	Args:    cobra.NoArgs,
}

var exerciseDownCmd = &cobra.Command{
	Use:   "down [exercise index]",
	Short: "Move the exercise down in the workout exercises",
	Long: `Move an exercise down in the workout exercises
  if the exercise is already the last exercise then nothing happens`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = editWorkout
		)

		e1, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("arg (%s) is not a digit", args[0])
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/down", username, workout, e1)
		handleResponse(doJSON(http.MethodPut, baseurl, endpoint, nil))
	},
}

var exerciseUpCmd = &cobra.Command{
	Use:   "up [exercise index]",
	Short: "Move the exercise up in the workout exercises",
	Long: `Move an exercise up in the workout exercises
  if the exercise is already the first exercise then nothing happens`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = editWorkout
		)

		e1, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("arg (%s) is not a digit", args[0])
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/up", username, workout, e1)
		handleResponse(doJSON(http.MethodPut, baseurl, endpoint, nil))
	},
}

var exerciseSwapCmd = &cobra.Command{
	Use:   "swap [exercise index] [exercise index]",
	Short: "swap two exercises with each other",
	Long: `swap the exercise provided by the first argument 
  with the exercise from the second argument. Use the indices
  of the index, you can find them using the list command.
  `,
	Args: cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = editWorkout
		)

		e1, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("arg (%s) is not a digit", args[0])
			os.Exit(1)
		}

		e2, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("arg (%s) is not a digit", args[1])
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/swap", username, workout, e1)
		body := strings.NewReader(fmt.Sprintf("{%q: %d}", "with", e2))

		handleResponse(doJSON(http.MethodPost, baseurl, endpoint, body))
	},
}
