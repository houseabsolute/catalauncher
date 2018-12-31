package config

type Config struct {
	rootDir string
}

func New(rootDir string) (*Config, error) {
	return &Config{rootDir: rootDir}, nil
}

func (c *Config) RootDir() string {
	return c.rootDir
}
