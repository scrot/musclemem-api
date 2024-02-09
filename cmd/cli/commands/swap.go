package commands

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(swapCmd)
}

var swapCmd = &cobra.Command{
	Use:   "swap [workout id] [exercise index] [exercise index]",
	Short: "swap two exercises with each other",
	Long: `swap the exercise provided by the first argument 
  with the exercise from the second argument. Use the indices
  of the index, you can find them using the list command.
  `,
	Args: cobra.ExactArgs(3),
	Run: func(_ *cobra.Command, args []string) {
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

		endpoint := fmt.Sprintf("/workouts/%d/exercises/%d/swap", wid, e1, e2)
		url, err := url.JoinPath(baseurl, endpoint)
		if err != nil {
			fmt.Printf("invalid url: %s", err)
			os.Exit(1)
		}

		body := fmt.Sprintf("{%q: %d}", "with", e2)
		req, err := http.NewRequest("POST", url, strings.NewReader(body))
		if err != nil {
			fmt.Printf("request err: %s", err)
			os.Exit(1)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("api err: %s", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("api err: %s", resp.Status)
			os.Exit(1)
		}

		fmt.Printf("exercise %d swapped with %d", e1, e2)
	},
}
