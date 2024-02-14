package register

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/spf13/cobra"
)

type RegisterOpts struct {
	UserFilePath string

	// TODO: add user without args?
	Username, Email, Password string
}

func NewRegisterCmd(config *cli.CLIConfig) *cobra.Command {
	opts := &RegisterOpts{}

	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Long:  `Create a new musclemem user`,
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				err  error
				data []byte
			)

			if opts.UserFilePath != "" {
				data, err = os.ReadFile(opts.UserFilePath)
				if err != nil {
					return cli.NewCLIError(err)
				}
			}

			if opts.Username != "" && opts.Email != "" && opts.Password != "" {
				user := user.User{
					Username: opts.Username,
					Email:    opts.Email,
					Password: opts.Password,
				}

				data, err = json.Marshal(user)
				if err != nil {
					return cli.NewCLIError(err)
				}
			}

			resp, err := cli.SendRequest(http.MethodPost, config.BaseURL, "/users", bytes.NewReader(data))
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				return cli.NewAPIStatusError(resp.StatusCode)
			}

			defer resp.Body.Close()
			dec := json.NewDecoder(resp.Body)

			var u user.User
			if err = dec.Decode(&u); err != nil {
				return cli.NewJSONError(err)
			}

			return nil
		},
	}

	registerCmd.Flags().StringVarP(&opts.UserFilePath, "file", "f", "", "path to json file (required)")
	registerCmd.Flags().StringVar(&opts.Username, "user", "", "username of user")
	registerCmd.Flags().StringVar(&opts.Email, "email", "", "email address of user")
	registerCmd.Flags().StringVar(&opts.Password, "password", "", "password of user")
	registerCmd.MarkFlagsRequiredTogether("user", "email", "password")

	return registerCmd
}
