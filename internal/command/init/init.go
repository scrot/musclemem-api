package init

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/scrot/musclemem-api/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type InitOptions struct {
	BaseURL    string
	ConfigPath string
	Overwrite  bool
}

func NewInitCmd(c *cli.CLIConfig) *cobra.Command {
	var opts InitOptions

	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialise cli",
		Long:  `setup new configuration for cli tool`,
		Example: heredoc.Doc(`
    $ mm init
    $ mm init --baseurl https://musclemem.app --config /path/to/configfile
    `),
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := NewConfigFile(c, opts.BaseURL, opts.Overwrite); err != nil {
				return cli.NewCLIError(err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.BaseURL, "baseurl", "https://musclemem.app", "url to musclemem-api server")
	cmd.Flags().BoolVar(&opts.Overwrite, "overwrite", false, "overwrite old configuration")

	return cmd
}

func NewConfigFile(c *cli.CLIConfig, base string, overwrite bool) error {
	configfile := cli.DefaultConfigPath(c.CLIConfigPath)

	// create new configfile if needed
	if _, err := os.Stat(configfile); err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if err := os.Mkdir(path.Dir(configfile), os.ModePerm); err != nil {
				return err
			}
			if _, err := os.Create(configfile); err != nil {
				return err
			}
		case errors.Is(err, fs.ErrExist):
			if overwrite {
				if _, err := os.Create(configfile); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("file already exists, use the overwrite flag")
			}
		default:
			return err
		}
	}

	viper.SetConfigFile(configfile)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.Set("baseurl", base)
	viper.Set("configpath", configfile)
	viper.Set("user", "")

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	// TODO: switch with Client...
	// c.BaseURL = base
	// c.CLIConfigPath = configfile

	return nil
}
