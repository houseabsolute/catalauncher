package config

import (
	"fmt"
	"path/filepath"
)

type Config struct {
	rootDir string
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
	// XXX - need to get "cataclysmdda-0.C" dynamically
	return filepath.Join(c.BuildsDir(), fmt.Sprintf("%d", num), "cataclysmdda-0.C")
}
