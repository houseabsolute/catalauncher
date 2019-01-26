package cmd

import (
	"github.com/houseabsolute/catalauncher/cleaner"
	"github.com/houseabsolute/catalauncher/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var max int
var keep []uint

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean out old builds",
	Long: `
The clean subcommand deletes old builds. By default it saves up to 5 builds
but you can override this with the "--max" flag. You can also keep specific
builds by passing the "--keep" flag.
`,
	Run: func(cmd *cobra.Command, args []string) {
		l, err := cleaner.New(viper.GetString("root"), max, keep)
		if err != nil {
			util.PrintErrorAndExit(err.Error())
		}

		err = l.Clean()
		if err != nil {
			util.PrintErrorAndExit(err.Error())
		}
	},
}

func init() {
	cleanCmd.PersistentFlags().IntVar(
		&max, "max", 5, "the max number of builds to keep")
	cleanCmd.PersistentFlags().UintSliceVar(
		&keep, "keep", []uint{}, "keep the specified build(s)")
	rootCmd.AddCommand(cleanCmd)
}
