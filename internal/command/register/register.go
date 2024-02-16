package register

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/spf13/cobra"
)

type RegisterOptions struct {
	user.User
	UserFilePath string
}

func NewRegisterCmd(config *cli.CLIConfig) *cobra.Command {
	opts := &RegisterOptions{}

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Long:  `Create a new musclemem user`,
		Args:  cobra.NoArgs,
		Example: heredoc.Doc(`
      $ mm register -f /path/to/user.json
      $ mm register --username anna --email anna@email.com --password passwd
    `),
		RunE: func(_ *cobra.Command, _ []string) error {
			var u user.User
			switch {
			case opts.UserFilePath != "":
				file, err := os.Open(opts.UserFilePath)
				if err != nil {
					return cli.NewCLIError(err)
				}

				dec := json.NewDecoder(file)
				if err := dec.Decode(&u); err != nil {
					return cli.NewJSONError(err)
				}
			case opts.Username != "" && opts.Password != "" && opts.Email != "":
				u = user.User{
					Username: opts.Username,
					Email:    opts.Email,
					Password: opts.Password,
				}
			default:
				return cli.NewCLIError(errors.New("missing flags"))

			}

			data, err := json.Marshal(u)
			if err != nil {
				return cli.NewCLIError(err)
			}

			resp, err := cli.SendRequest(http.MethodPost, config.BaseURL, "/users", bytes.NewReader(data))
			if err != nil {
				return cli.NewAPIError(err)
			}

			if resp.StatusCode != http.StatusOK {
				switch resp.StatusCode {
				case http.StatusConflict:
					return cli.NewAPIError(cli.ErrExists)
				default:
					return cli.NewAPIStatusError(resp)
				}
			}

			defer resp.Body.Close()
			dec := json.NewDecoder(resp.Body)

			var nu user.User
			if err = dec.Decode(&nu); err != nil {
				return cli.NewJSONError(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.UserFilePath, "file", "f", "", "path to json file")
	cmd.Flags().StringVar(&opts.Username, "username", "", "username of user")
	cmd.Flags().StringVar(&opts.Email, "email", "", "email address of user")
	cmd.Flags().StringVar(&opts.Password, "password", "", "password of user")
	cmd.MarkFlagsRequiredTogether("username", "email", "password")
	cmd.MarkFlagsMutuallyExclusive("username", "file")
	cmd.MarkFlagsOneRequired("username", "file")

	return cmd
}
