package register

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	model "github.com/scrot/musclemem-api/internal/user"
	"github.com/spf13/cobra"
)

type RegisterOptions struct {
	model.User
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
    $ mm register --username anna --email anna@email.com --password passwd
    `),
		RunE: func(_ *cobra.Command, _ []string) error {
			user := model.User{
				Username: opts.Username,
				Email:    opts.Email,
				Password: opts.Password,
			}

			data, err := json.Marshal(user)
			if err != nil {
				return cli.NewCLIError(err)
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

			var u model.User
			if err = dec.Decode(&u); err != nil {
				return cli.NewJSONError(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Username, "user", "", "username of user")
	cmd.Flags().StringVar(&opts.Email, "email", "", "email address of user")
	cmd.Flags().StringVar(&opts.Password, "password", "", "password of user")
	cmd.MarkFlagsRequiredTogether("user", "email", "password")

	return cmd
}
