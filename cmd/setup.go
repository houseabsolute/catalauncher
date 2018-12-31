package cmd

import (
	"github.com/houseabsolute/catalauncher/setupper"
	"github.com/houseabsolute/catalauncher/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run this to set the launcher up",
	Long: `This command will run you through some prompts to set up catalauncher.

You must run this at least once before running any other command. You can also
run it later to change your setup.
`,
	Run: func(cmd *cobra.Command, args []string) {
		s, err := setupper.New(viper.GetString("root"), cmd.Flag("config").Value.String())
		if err != nil {
			util.PrintErrorAndExit(err.Error())
		}

		err = s.Setup()
		if err != nil {
			util.PrintErrorAndExit(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
