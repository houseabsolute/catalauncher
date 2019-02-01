package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/houseabsolute/catalauncher/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "catalauncher",
	Short: "Manage and launch Cataclysm: Dark Days Ahead",
	Long: `This is a tool for managing and launching Cataclysm: Dark Days Ahead.

It will download new versions of the game, install mods, tilesets, and
soundpacks, and even let you save scum if you want.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		util.PrintErrorAndExit(err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "", "config file (default is ~/.catalauncher/config.toml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			util.PrintErrorAndExit("Could not find your home directory: %s", err)
		}

		viper.AddConfigPath(filepath.Join(home, ".catalauncher"))
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
