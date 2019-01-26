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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
