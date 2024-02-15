package info

import (
	"fmt"

	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
)

type InfoOptions struct{}

func NewInfoCmd(_ *cli.CLIConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "cli information",
		Long:  "prints information about the client",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return fmt.Errorf("not implemented")
		},
	}

	return cmd
}
