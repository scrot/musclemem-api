package login

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

func init() {
}

type LoginOptions struct {
	Username, Password string
}

func NewLoginCmd(c *cli.CLIConfig) *cobra.Command {
	var opts LoginOptions

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in the user",
		Long: `Binds the cli tool to a specific user, all
    all subsequent actions will be done as if by the user`,
		Example: heredoc.Doc(`
      $ mm login --username anne@email.com --password passwd
    `),
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if c.User != "" {
				err := fmt.Errorf("already logged-in, you need to logout first")
				return cli.NewCLIError(err)
			}

			// TODO: exchange for api-token?

			if err := keyring.Set(c.CLIName, opts.Username, opts.Password); err != nil {
				return cli.NewCLIError(err)
			}

			viper.Set("user", opts.Username)
			if err := viper.WriteConfig(); err != nil {
				return cli.NewCLIError(err)
			}
			c.User = opts.Username

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Username, "username", "", "username of user")
	cmd.MarkFlagRequired("user")
	cmd.Flags().StringVar(&opts.Password, "password", "", "password of user")
	cmd.MarkFlagRequired("password")

	return cmd
}
