package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Config struct {
	rootDir string
	gameDir string
}

func New(rootDir string) (*Config, error) {
	return &Config{rootDir: rootDir}, nil
}

func (c *Config) RootDir() string {
	return c.rootDir
}

func (c *Config) GameDataDir() string {
	return filepath.Join(c.RootDir(), "game-data")
}

func (c *Config) ExtrasDir() string {
	return filepath.Join(c.RootDir(), "extras")
}

func (c *Config) BuildsDir() string {
	return filepath.Join(c.RootDir(), "builds")
}

func (c *Config) GameDir(num uint) string {
	if c.gameDir != "" {
		return c.gameDir
	}

	root := filepath.Join(c.BuildsDir(), fmt.Sprintf("%d", num))
	entries, err := ioutil.ReadDir(root)
	if err != nil {
		// XXX - This should be returned but there's a bunch of places to
		// change.
		panic(err)
	}
	var dir string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "cataclysmdda-0.E") {
			dir = e.Name()
			break
		}
	}
	if dir == "" {
		panic(fmt.Sprintf("Could not find cataclysmdda-? dir in %s", root))
	}

	c.gameDir = filepath.Join(root, dir)

	return c.gameDir
}
