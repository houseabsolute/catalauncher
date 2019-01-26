package cmd

import (
	"github.com/houseabsolute/catalauncher/launcher"
	"github.com/houseabsolute/catalauncher/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var build uint

// launchCmd represents the launch command
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch Cataclysm: DDA",
	Long: `
The launch subcommand will start Cataclysm: DDA in a Docker container, keeping
your saves and config in a directory on your host machine. By default it
always downloads the latest build but you can override that with the "--build"
flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		l, err := launcher.New(viper.GetString("root"), build)
		if err != nil {
			util.PrintErrorAndExit(err.Error())
		}

		err = l.Launch()
		if err != nil {
			util.PrintErrorAndExit(err.Error())
		}
	},
}

func init() {
	launchCmd.PersistentFlags().UintVar(
		&build, "build", 0, "the build number to launch (defaults to the latest)")
	rootCmd.AddCommand(launchCmd)
}
