package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/scrot/musclemem-api/internal/exercise"
	"github.com/scrot/musclemem-api/internal/workout"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	editExerciseCmd.Flags().StringVar(&changeExerciseName, "name", "", "change exercise name")
	editExerciseCmd.Flags().Float64Var(&changeExerciseWeight, "weight", 0, "change exercise weight")
	editExerciseCmd.Flags().IntVar(&changeExerciseRepetitions, "reps", 0.0, "change exercise repetitons")
	editExerciseCmd.PersistentFlags().IntVarP(&changeWorkoutExercise, "workout", "w", 0, "index of user workout (required)")
	editExerciseCmd.MarkPersistentFlagRequired("workout")
	editExerciseCmd.AddCommand(exerciseDownCmd, exerciseUpCmd, exerciseSwapCmd)

	editWorkoutCmd.Flags().StringVarP(&changeWorkoutName, "name", "n", "", "change workout name")

	editCmd.AddCommand(editExerciseCmd, editWorkoutCmd)

	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an user resource",
	Long:  `Edit an user resource, the user must be logged in`,
	Args:  cobra.NoArgs,
}

var (
	changeWorkoutExercise     int
	changeExerciseName        string
	changeExerciseWeight      float64
	changeExerciseRepetitions int

	editExerciseCmd = &cobra.Command{
		Use:     "exercise INDEX",
		Aliases: []string{"ex"},
		Short:   "Edit an exercise",
		Long:    `Edit an exercise of a user workout, the user must be logged in`,
		Args:    cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			username := viper.GetString("user")

			if changeExerciseName == "" && changeExerciseWeight == 0 && changeExerciseRepetitions == 0 {
				handleCLIError(errors.New("missing edit flags"))
			}

			ei, err := strconv.Atoi(args[0])
			if err != nil {
				handleCLIError(err)
			}

			patch := exercise.Exercise{
				Name:        changeExerciseName,
				Weight:      changeExerciseWeight,
				Repetitions: changeExerciseRepetitions,
			}

			payload, err := json.Marshal(&patch)
			if err != nil {
				handleCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d", username, changeWorkoutExercise, ei)
			handleResponse(doJSON("PATCH", baseurl, endpoint, bytes.NewReader(payload)))
		},
	}
)

var (
	changeWorkoutName string
	editWorkoutCmd    = &cobra.Command{
		Use:     "workout INDEX",
		Aliases: []string{"wo"},
		Short:   "Edit an workout",
		Long:    `Edit an user workout, the user must be logged in`,
		Args:    cobra.MaximumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			var (
				username = viper.GetString("user")
				wo       = args[0]
				name     = changeWorkoutName
			)

			if changeWorkoutName == "" {
				handleCLIError(errors.New("missing edit flag"))
			}

			patch := workout.Workout{Name: name}
			payload, err := json.Marshal(patch)
			if err != nil {
				handleCLIError(err)
			}

			endpoint := fmt.Sprintf("/users/%s/workouts/%s", username, wo)
			handleResponse(doJSON("PATCH", baseurl, endpoint, bytes.NewReader(payload)))
		},
	}
)

var exerciseDownCmd = &cobra.Command{
	Use:   "down INDEX",
	Short: "Move the exercise down in the workout exercises",
	Long: `Move an exercise down in the workout exercises
  if the exercise is already the last exercise then nothing happens`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = changeWorkoutExercise
		)

		e1, err := strconv.Atoi(args[0])
		if err != nil {
			handleCLIError(errors.New("invalid arg"))
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/down", username, workout, e1)
		handleResponse(doJSON(http.MethodPut, baseurl, endpoint, nil))
	},
}

var exerciseUpCmd = &cobra.Command{
	Use:   "up INDEX",
	Short: "Move the exercise up in the workout exercises",
	Long: `Move an exercise up in the workout exercises
  if the exercise is already the first exercise then nothing happens`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = changeWorkoutExercise
		)

		e1, err := strconv.Atoi(args[0])
		if err != nil {
			handleCLIError(errors.New("invalid args"))
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/up", username, workout, e1)
		handleResponse(doJSON(http.MethodPut, baseurl, endpoint, nil))
	},
}

var exerciseSwapCmd = &cobra.Command{
	Use:   "swap INDEX1 INDEX2",
	Short: "swap two exercises with each other",
	Long: `swap the exercise provided by the first argument 
  with the exercise from the second argument. Use the indices
  of the index, you can find them using the list command.
  `,
	Args: cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		var (
			username = viper.GetString("user")
			workout  = changeWorkoutExercise
		)

		e1, err := strconv.Atoi(args[0])
		if err != nil {
			handleCLIError(errors.New("invalid args"))
		}

		e2, err := strconv.Atoi(args[1])
		if err != nil {
			handleCLIError(errors.New("invalid args"))
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/swap", username, workout, e1)
		body := strings.NewReader(fmt.Sprintf("{%q: %d}", "with", e2))

		handleResponse(doJSON(http.MethodPost, baseurl, endpoint, body))
	},
}
