package command

import (
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/command/add"
	"github.com/scrot/musclemem-api/internal/command/edit"
	"github.com/scrot/musclemem-api/internal/command/info"
	ini "github.com/scrot/musclemem-api/internal/command/init"
	"github.com/scrot/musclemem-api/internal/command/list"
	"github.com/scrot/musclemem-api/internal/command/login"
	"github.com/scrot/musclemem-api/internal/command/logout"
	"github.com/scrot/musclemem-api/internal/command/move"
	"github.com/scrot/musclemem-api/internal/command/register"
	"github.com/scrot/musclemem-api/internal/command/remove"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

type RootOptions struct{}

func NewRootCmd(c *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   c.CLIName,
		Short: "A cli tool for interacting with the musclemem-api",
		Long: `Musclemem is a simple fitness routine application
  structuring workout exercises and tracking performance`,
		Version: c.CLIVersion,
		Example: heredoc.Doc(`
			$ mm login
			$ mm add exercise -w 1 
			$ mm edit workout --name "workout 1"
		`),
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(
		add.NewAddCmd(c),
		remove.NewRemoveCmd(c),
		list.NewListCmd(c),
		edit.NewEditCmd(c),
		move.NewMoveCmd(c),
		login.NewLoginCmd(c),
		logout.NewLogoutCmd(c),
		register.NewRegisterCmd(c),
		info.NewInfoCmd(c),
		ini.NewInitCmd(c),
	)

	return cmd
}
