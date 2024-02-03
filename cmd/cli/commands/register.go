package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/scrot/musclemem-api/internal/user"
	"github.com/spf13/cobra"
)

func init() {
	registerCmd.Flags().StringVarP(&newUsername, "user", "u", "", "username of user")
	registerCmd.MarkFlagRequired("user")
	registerCmd.Flags().StringVarP(&newEmail, "email", "e", "", "email address of user")
	registerCmd.MarkFlagRequired("email")
	registerCmd.Flags().StringVarP(&newPassword, "password", "p", "", "password of user")
	registerCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(registerCmd)
}

var (
	newUsername string
	newEmail    string
	newPassword string
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  `Create a new musclemem user`,
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		user := user.User{
			Username: newUsername,
			Email:    newEmail,
			Password: newPassword,
		}

		payload, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resp, err := postJSON(baseurl, "/users", bytes.NewReader(payload))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		body, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("api error: %s\n", body)
			os.Exit(1)
		}

		fmt.Printf("Registered new user with id %s\n", body)
	},
}
