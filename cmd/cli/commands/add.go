package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	addCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "path to json file (required)")
	addCmd.MarkPersistentFlagRequired("file")

	addCmd.AddCommand(addWorkoutCmd)
	addCmd.AddCommand(addExerciseCmd)

	rootCmd.AddCommand(addCmd)
}

var filePath string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new workouts or exercises",
	Long: `Allows the creation of new user workouts
  and new workout exercises`,
}

var addWorkoutCmd = &cobra.Command{
	Use:     "workout",
	Aliases: []string{"wo"},
	Short:   "Add new workout",
	Long:    `Add a new workout to the currently signed in user`,
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		endpoint := fmt.Sprintf("/users/%s/workouts", viper.GetString("user"))
		resp, err := postJSON(baseurl, endpoint, file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("api error: %s", resp.Status)
		}
	},
}

var addExerciseCmd = &cobra.Command{
	Use:     "exercise [workout index]",
	Aliases: []string{"ex"},
	Short:   "Add new exercise",
	Long:    `Add a new exercise to the workout for the signed in user`,
	Args:    cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		var (
			user    = viper.GetString("user")
			workout = args[0]
		)

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()

		endpoint := fmt.Sprintf("/users/%s/workouts/%s/exercises", user, workout)
		if _, err := postJSON(baseurl, endpoint, file); err != nil {
			fmt.Println(err)
		}
	},
}
