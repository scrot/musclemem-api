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
	rootCmd.AddCommand(downCmd, upCmd, swapCmd)
}

var downCmd = &cobra.Command{
	Use:   "down [workout index] [exercise index]",
	Short: "move the exercise down in the workout exercises",
	Long: `move an exercise down in the workout exercises
  if the exercise is already the last exercise then nothing happens`,
	Args: cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		username := viper.GetString("user")

		wid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("1st arg (%s) is not a digit", args[0])
			os.Exit(1)
		}

		e1, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("2nd arg (%s) is not a digit", args[1])
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/down", username, wid, e1)
		resp, err := doJSON(http.MethodPut, baseurl, endpoint, nil)
		if err != nil {
			fmt.Printf("api error: %s\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("api error: %s\n", resp.Status)
			os.Exit(1)
		}
	},
}

var upCmd = &cobra.Command{
	Use:   "up [workout index] [exercise index]",
	Short: "move the exercise up in the workout exercises",
	Long: `move an exercise up in the workout exercises
  if the exercise is already the first exercise then nothing happens`,
	Args: cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		username := viper.GetString("user")

		wid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("1st arg (%s) is not a digit", args[0])
			os.Exit(1)
		}

		e1, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("2nd arg (%s) is not a digit", args[1])
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/up", username, wid, e1)
		resp, err := doJSON(http.MethodPut, baseurl, endpoint, nil)
		if err != nil {
			fmt.Printf("api error: %s\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("api error: %s\n", resp.Status)
			os.Exit(1)
		}
	},
}

var swapCmd = &cobra.Command{
	Use:   "swap [workout index] [exercise index] [exercise index]",
	Short: "swap two exercises with each other",
	Long: `swap the exercise provided by the first argument 
  with the exercise from the second argument. Use the indices
  of the index, you can find them using the list command.
  `,
	Args: cobra.ExactArgs(3),
	Run: func(_ *cobra.Command, args []string) {
		username := viper.GetString("user")

		wid, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("1st arg (%s) is not a digit", args[0])
			os.Exit(1)
		}

		e1, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("2nd arg (%s) is not a digit", args[1])
			os.Exit(1)
		}

		e2, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("3rd arg (%s) is not a digit", args[2])
			os.Exit(1)
		}

		endpoint := fmt.Sprintf("/users/%s/workouts/%d/exercises/%d/swap", username, wid, e1)
		body := strings.NewReader(fmt.Sprintf("{%q: %d}", "with", e2))

		resp, err := doJSON(http.MethodPost, baseurl, endpoint, body)
		if err != nil {
			fmt.Printf("api error: %s\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("api error: %s\n", resp.Status)
			os.Exit(1)
		}
	},
}
