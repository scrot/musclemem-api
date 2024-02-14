package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logs out the current user",
	Long:  `Remove the credentials of the logged in user`,
	Run: func(_ *cobra.Command, _ []string) {
		if viper.GetString("user") == "" {
			fmt.Println("no user logged in")
			return
		}

		if err := keyring.Delete(appname, viper.GetString("user")); err != nil {
			handleCLIError(err)
		}

		username = ""
		viper.Set("user", username)
		if err := viper.WriteConfig(); err != nil {
			handleCLIError(err)
		}

		fmt.Println("logged out")
	},
}
