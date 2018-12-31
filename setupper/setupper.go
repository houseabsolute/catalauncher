package setupper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/houseabsolute/catalauncher/curuser"
	"github.com/houseabsolute/catalauncher/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

type Setupper struct {
	rootDir    string
	configFile string
	user       *curuser.User
}

func New(rootDir, configFile string) (*Setupper, error) {
	user, err := curuser.New()
	if err != nil {
		return nil, err
	}

	return &Setupper{rootDir, configFile, user}, nil
}

func (s *Setupper) Setup() error {
	err := s.checkForDocker()
	if err != nil {
		return err
	}

	err = s.setRootDirectory()
	if err != nil {
		return err
	}

	return nil
}

func (s *Setupper) checkForDocker() error {
	_, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("To use catalauncher you must have docker in your $PATH: %s", err)
	}
	return nil
}

func (s *Setupper) setRootDirectory() error {
	def := s.defaultRoot()
	rootDir, err := promptFor("Catalauncher root dir (stores game, saves, mods, etc.)?", def)
	if err != nil {
		return err
	}

	err = s.makeRoot(rootDir)
	if err != nil {
		return err
	}

	viper.Set("root", rootDir)
	file := filepath.Join(rootDir, "config.toml")

	util.Say(os.Stdout, "Writing your config file at %s", file)
	err = viper.WriteConfigAs(file)
	if err != nil {
		return fmt.Errorf("Could not write your config file: %s", err)
	}

	return nil
}

func (s *Setupper) defaultRoot() string {
	if s.rootDir != "" {
		return s.rootDir
	} else if s.configFile != "" {
		return filepath.Dir(s.configFile)
	}
	return filepath.Join(s.user.HomeDir, ".catalauncher")
}

var promptTemplates = &promptui.PromptTemplates{
	Prompt:  "{{ . }} ",
	Valid:   "{{ . | green }} ",
	Invalid: "{{ . | red }} ",
	Success: "{{ . | bold }} ",
}

func promptFor(label, def string) (string, error) {
	prompt := promptui.Prompt{
		Label:     label + " ",
		Default:   def,
		AllowEdit: true,
		Templates: promptTemplates,
	}
	val, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("Error attempting to run a prompt: %s", err)
	}
	return val, nil
}

func (s *Setupper) makeRoot(rootDir string) error {
	rootDir = strings.Replace(rootDir, "$HOME", s.user.HomeDir, -1)
	rootDir = strings.Replace(rootDir, "~", s.user.HomeDir, -1)

	exists, err := util.PathExists(rootDir)
	if err != nil {
		return fmt.Errorf("Could not check if the path at %s exists: %s", rootDir, err)
	}

	if !exists {
		util.Say(os.Stdout, "Creating the data rootDir at %s", rootDir)
		err := os.MkdirAll(rootDir, 0755)
		if err != nil {
			return fmt.Errorf("Error creating a directory at %s: %s", rootDir, err)
		}
	}

	return nil
}
