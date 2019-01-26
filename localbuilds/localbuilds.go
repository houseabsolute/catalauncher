package localbuilds

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/houseabsolute/catalauncher/config"
)

type LocalBuilds struct {
	config *config.Config
	builds *[]uint
}

func New(c *config.Config) *LocalBuilds {
	return &LocalBuilds{config: c}
}

func (l *LocalBuilds) Latest() (uint, error) {
	local, err := l.All()
	if err != nil {
		return 0, err
	}

	if len(local) == 0 {
		return 0, nil
	}

	return local[len(local)-1], nil
}

var buildNumberRE = regexp.MustCompile(`^[1-9][0-9]*$`)

func (l *LocalBuilds) All() ([]uint, error) {
	if l.builds != nil {
		return *l.builds, nil
	}

	local := []uint{}

	files, err := ioutil.ReadDir(l.config.BuildsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return local, nil
		}
		return local, fmt.Errorf("Could not read directory at %s: %s", l.config.BuildsDir(), err)
	}

	for _, f := range files {
		if f.IsDir() && buildNumberRE.MatchString(f.Name()) {
			i, err := strconv.Atoi(f.Name())
			if err != nil {
				return local, fmt.Errorf("Could not convert %s to an integer: %s", f.Name(), err)
			}
			local = append(local, uint(i))
		}
	}

	sort.Slice(local, func(i, j int) bool { return local[i] < local[j] })
	l.builds = &local

	return local, nil
}

func (l *LocalBuilds) HasBuild(wanted uint) (bool, error) {
	all, err := l.All()
	if err != nil {
		return false, err
	}

	for _, l := range all {
		if l == wanted {
			return true, nil
		}
	}
	return false, nil
}
