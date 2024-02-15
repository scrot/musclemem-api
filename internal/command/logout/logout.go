package logout

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

func NewLogoutCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout the current user",
		Long: `Remove the credentials from the system 
    and unbinds the user from the cli tool.`,
		Example: heredoc.Doc(`
    $ mm logout
    `),
		RunE: func(_ *cobra.Command, _ []string) error {
			if c.User == "" {
				return fmt.Errorf("no user logged in")
			}

			if err := keyring.Delete(c.CLIName, c.User); err != nil {
				return cli.NewCLIError(err)
			}

			viper.Set("user", "")
			if err := viper.WriteConfig(); err != nil {
				return cli.NewCLIError(err)
			}
			c.User = ""

			return nil
		},
	}

	return cmd
}
