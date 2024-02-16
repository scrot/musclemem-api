package cli

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

// CLIConfig contains configuration data for cli commands
type CLIConfig struct {
	// User is the currently logged-in user, all commands
	// are performed through this user, if it is empty then
	// no user is currently logged-in
	User string

	// BaseURL is the url to the musclemem-api server
	// all commands use this url to interact with
	BaseURL string

	// CLIDate is the date the cli tool was build
	CLIDate string

	// CLIAuthor is the author whom build the cli tool
	CLIAuthor string

	// CLIVersion is the build version of the cli tool
	CLIVersion string

	// CLIConfigPath is the path to the cli configuration file
	CLIConfigPath string

	// CLIName is the name of the cli tool and is used
	// to store configuration files under
	CLIName string

	// Out is the default output stream to write to
	Out io.Writer

	// Out is the output stream to write error output to
	OutErr io.Writer
}

// NewCLIConfig creates a new CLIConfig
func NewCLIConfig(name, version, author, date string) (*CLIConfig, error) {
	configpath := DefaultConfigPath(name)
	viper.SetConfigFile(configpath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &CLIConfig{
		CLIAuthor:  author,
		CLIVersion: version,
		CLIDate:    date,
		CLIName:    name,

		User:    viper.GetString("user"),
		BaseURL: viper.GetString("baseurl"),

		Out:    os.Stdout,
		OutErr: os.Stderr,
	}

	return config, nil
}

func DefaultConfigPath(appname string) string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(NewCLIError(err))
		os.Exit(1)
	}
	configfile := path.Join(home, "."+appname, "config.yaml")
	return configfile
}

func (c *CLIConfig) UserPassword() (string, error) {
	return keyring.Get(c.CLIName, c.User)
}
