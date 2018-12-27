package cmd

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/houseabsolute/catalauncher/util"
	"github.com/manifoldco/promptui"
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
		setup(cmd)
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func setup(cmd *cobra.Command) {
	u := currentUser()
	def := defaultRoot(u, cmd)
	rootDir := promptFor("Root directory (stores game, saves, mods, etc.)?", def)

	makeRoot(u, rootDir)

	viper.Set("root", rootDir)
	file := filepath.Join(rootDir, "config.toml")

	util.Say(os.Stdout, "Writing your config file at %s", file)
	err := viper.WriteConfigAs(file)
	if err != nil {
		printErrorAndExit("Could not write your config file: %s", err)
	}
}

func currentUser() *user.User {
	u, err := user.Current()
	if err != nil {
		printErrorAndExit("Could not get get the current user: %s", err)
	}
	return u
}

func defaultRoot(u *user.User, cmd *cobra.Command) string {
	if c := viper.GetString("root"); c != "" {
		return c
	} else if c := cmd.Flag("config").Value.String(); c != "" {
		return filepath.Dir(c)
	}
	return filepath.Join(u.HomeDir, ".catalauncher")
}

var promptTemplates = &promptui.PromptTemplates{
	Prompt:  "{{ . }} ",
	Valid:   "{{ . | green }} ",
	Invalid: "{{ . | red }} ",
	Success: "{{ . | bold }} ",
}

func promptFor(label, def string) string {
	prompt := promptui.Prompt{
		Label:     label + " ",
		Default:   def,
		AllowEdit: true,
		Templates: promptTemplates,
	}
	val, err := prompt.Run()
	if err != nil {
		printErrorAndExit("Error attempting to run a prompt: %s", err)
	}
	return val
}

func makeRoot(u *user.User, rootDir string) {
	rootDir = strings.Replace(rootDir, "$HOME", u.HomeDir, -1)
	rootDir = strings.Replace(rootDir, "~", u.HomeDir, -1)

	exists, err := util.PathExists(rootDir)
	if err != nil {
		printErrorAndExit("Could not check if the path at %s exists: %s", rootDir, err)
	}

	if !exists {
		util.Say(os.Stdout, "Creating the data rootDir at %s", rootDir)
		os.MkdirAll(rootDir, 0755)
	}
}

func printErrorAndExit(tmpl string, args ...interface{}) {
	util.Say(os.Stderr, tmpl, args)
}
