package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

func init() {
	loginCmd.Flags().StringVarP(&email, "user", "u", "", "email address of user")
	loginCmd.MarkFlagRequired("user")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "password of user")
	loginCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(loginCmd)
}

var (
	email    string
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
			fmt.Println("already logged-in, you need to logout first")
			os.Exit(1)
		}

		if err := keyring.Set(appname, email, password); err != nil {
			fmt.Printf("storing credentials in keyring: %s", err)
			os.Exit(1)
		}

		viper.Set("user", email)
		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("updating configuration file: %s", err)
			os.Exit(1)
		}

		fmt.Println("logged in")
	},
}
