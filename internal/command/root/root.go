package command

import (
	"fmt"

	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/scrot/musclemem-api/internal/command/add"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

type RootOptions struct {
	ConfigPath string
}

func NewRootCmd(c *cli.CLIConfig) *cobra.Command {
	opts := RootOptions{}

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

	description := fmt.Sprintf("config file (default is %s)", c.CLIConfigPath)
	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", description)

	cmd.AddCommand(
		add.NewAddCmd(c),
	)

	return cmd
}
