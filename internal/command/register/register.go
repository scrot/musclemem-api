package register

import (
	"context"
	"encoding/json"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/user"
	"github.com/spf13/cobra"
)

type RegisterOptions struct {
	User user.User

	UserFilePath string
}

func NewRegisterCmd(c *cli.CLIConfig) *cobra.Command {
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
			if opts.UserFilePath != "" {
				file, err := os.Open(opts.UserFilePath)
				if err != nil {
					return cli.NewCLIError(err)
				}

				dec := json.NewDecoder(file)
				if err := dec.Decode(&opts.User); err != nil {
					return cli.NewCLIError(err)
				}
			}

			if _, _, err := c.Users.Register(context.TODO(), opts.User); err != nil {
				return cli.NewAPIError(err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.UserFilePath, "file", "f", "", "path to json file")
	cmd.Flags().StringVar(&opts.User.Username, "username", "", "username of user")
	cmd.Flags().StringVar(&opts.User.Email, "email", "", "email address of user")
	cmd.Flags().BytesHexVar(&opts.User.Password, "password", []byte{}, "password of user")
	cmd.MarkFlagsRequiredTogether("username", "email", "password")
	cmd.MarkFlagsMutuallyExclusive("username", "file")
	cmd.MarkFlagsOneRequired("username", "file")

	return cmd
}
