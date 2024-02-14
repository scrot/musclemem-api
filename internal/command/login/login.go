package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

func init() {
	loginCmd.Flags().StringVarP(&username, "user", "u", "", "email address of user")
	loginCmd.MarkFlagRequired("user")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "password of user")
	loginCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(loginCmd)
}

var (
	username string
	password string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Logs in the user given the credentials",
	Long: `Binds musclemem to a specific user, all
  all subsequent actions will be done as if by the user`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if viper.GetString("user") != "" {
			err := fmt.Errorf("already logged-in, you need to logout first")
			handleCLIError(err)
		}

		if err := keyring.Set(appname, username, password); err != nil {
			handleCLIError(err)
		}

		viper.Set("user", username)
		if err := viper.WriteConfig(); err != nil {
			handleCLIError(err)
		}
	},
}
