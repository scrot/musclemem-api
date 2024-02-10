package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/scrot/musclemem-api/internal/user"
	"github.com/spf13/cobra"
)

func init() {
	registerCmd.Flags().StringVarP(&userFilePath, "file", "f", "", "path to json file (required)")
	registerCmd.Flags().StringVarP(&newUsername, "user", "u", "", "username of user")
	registerCmd.Flags().StringVarP(&newEmail, "email", "e", "", "email address of user")
	registerCmd.Flags().StringVarP(&newPassword, "password", "p", "", "password of user")
	registerCmd.MarkFlagsRequiredTogether("user", "email", "password")

	rootCmd.AddCommand(registerCmd)
}

var (
	userFilePath string

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
		var (
			err  error
			data []byte
		)

		if userFilePath != "" {
			data, err = os.ReadFile(userFilePath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			user := user.User{
				Username: newUsername,
				Email:    newEmail,
				Password: newPassword,
			}

			data, err = json.Marshal(user)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		resp, err := doJSON(http.MethodPost, baseurl, "/users", bytes.NewReader(data))
		handleResponse(resp, err)

		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)

		var u user.User
		if err = dec.Decode(&u); err != nil {
			fmt.Printf("decode error: %s\n", err)
			os.Exit(1)
		}
	},
}
