package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/zalando/go-keyring"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version    = "0.0.1"
	appname    = "musclemem"
	baseurl    = "http://localhost:8080"
	configfile = ".musclemem/config.yaml"
	config     string
)

func init() {
	cobra.OnInitialize(loadConfig)
	cobra.OnInitialize(loadCredentials)

	rootCmd.PersistentFlags().StringVar(&config, "config", "", "config file (default is $HOME/.cobra.yaml)")
}

var rootCmd = &cobra.Command{
	Use:   "musclemem",
	Short: "A cli tool for interacting with the musclemem-api",
	Long: `Musclemem is a simple fitness routine application
  structuring workout exercises and tracking performance`,
	Version: version,
	Run:     func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadConfig() {
	if config == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		config = path.Join(home, configfile)
	}

	if _, err := os.Stat(config); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			os.Mkdir(path.Dir(config), os.ModePerm)
			os.Create(config)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	viper.SetConfigFile(config)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("unable to read config file: %s", err)
		os.Exit(1)
	}
}

func loadCredentials() {
	user := viper.GetString("user")
	if user != "" {
		var err error
		if password, err = keyring.Get(appname, user); err != nil {
			fmt.Printf("unable to retreive %s/%s credentials from keyring: %s",
				appname, user, err)
			os.Exit(1)
		}
	}
}
