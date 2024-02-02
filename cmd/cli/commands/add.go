package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	addCmd.PersistentFlags().StringVarP(&filepath, "file", "f", "", "path to json file (required)")
	addCmd.MarkPersistentFlagRequired("file")

	addCmd.AddCommand(addWorkoutCmd)
	addCmd.AddCommand(addExerciseCmd)

	rootCmd.AddCommand(addCmd)
}

var (
	filepath string
	addCmd   = &cobra.Command{
		Use:   "add",
		Short: "Add new workouts or exercises",
		Long: `Allows the creation of new user workouts
  and new workout exercises`,
	}
)

var addWorkoutCmd = &cobra.Command{
	Use:     "workout",
	Aliases: []string{"wo"},
	Short:   "Add new workout",
	Long:    `Add a new workout to the currently signed in user`,
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		if _, err := postJSON(baseurl, "/workouts", file); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var addExerciseCmd = &cobra.Command{
	Use:     "exercise",
	Aliases: []string{"ex"},
	Short:   "Add new exercise",
	Long:    `Add a new exercise to the workout for the signed in user`,
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		if _, err := postJSON(baseurl, "/exercises", file); err != nil {
			fmt.Println(err)
		}
	},
}
