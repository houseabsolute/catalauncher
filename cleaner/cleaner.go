package cleaner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/houseabsolute/catalauncher/config"
	"github.com/houseabsolute/catalauncher/localbuilds"
	"github.com/houseabsolute/catalauncher/util"
)

type Cleaner struct {
	config *config.Config
	local  *localbuilds.LocalBuilds
	max    int
	keep   []uint
	stdout io.Writer
	stderr io.Writer
}

func New(rootDir string, max int, keep []uint) (*Cleaner, error) {
	c, err := config.New(rootDir)
	if err != nil {
		return nil, err
	}

	return &Cleaner{
		config: c,
		local:  localbuilds.New(c),
		max:    max,
		keep:   keep,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}, nil
}

func (c *Cleaner) Clean() error {
	all, err := c.local.All()
	if err != nil {
		return err
	}

	if len(all) <= c.max {
		util.Say(c.stdout, "Keeping all %d builds so there is nothing to clean", len(all))
		return nil
	}

	shouldKeep := map[uint]bool{}
	for _, k := range c.keep {
		shouldKeep[k] = true
	}

	rest := all[0 : len(all)-c.max]

	for _, b := range rest {
		if shouldKeep[b] {
			util.Say(c.stdout, "Keeping build %d as requested", b)
		} else {
			util.Say(c.stdout, "Deleting build %d", b)
			err := os.RemoveAll(filepath.Join(c.config.BuildsDir(), fmt.Sprintf("%d", b)))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
