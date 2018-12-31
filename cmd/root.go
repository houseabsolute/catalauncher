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

// rootCmd represents the base command when called without any subcommands
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		util.PrintErrorAndExit(err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.catalauncher/config.yaml)")
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
