package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "cli information",
	Long:  "prints configuration information about the client",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("not implemented")
	},
}
