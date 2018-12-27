package cmd

import (
	"github.com/houseabsolute/catalauncher/launcher"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
		launcher.New(viper.GetString("root")).Launch()
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}
