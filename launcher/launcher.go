package launcher

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/houseabsolute/catalauncher/config"
	"github.com/houseabsolute/catalauncher/curuser"
	"github.com/houseabsolute/catalauncher/util"
	pb "gopkg.in/cheggaaa/pb.v2"
)

type build struct {
	uri         string
	filename    string
	version     string
	buildNumber uint
	date        time.Time
}

type Launcher struct {
	config      *config.Config
	build       uint
	user        *curuser.User
	stdout      io.Writer
	stderr      io.Writer
	buildsURI   string
	currentUser *user.User
}

const defaultBuildsURI = "http://dev.narc.ro/cataclysm/jenkins-latest/Linux_x64/Tiles/"

func New(rootDir string, build uint) (*Launcher, error) {
	c, err := config.New(rootDir)
	if err != nil {
		return nil, err
	}

	user, err := curuser.New()
	if err != nil {
		return nil, err
	}

	return &Launcher{
		config:    c,
		build:     build,
		user:      user,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		buildsURI: defaultBuildsURI,
	}, nil
}

func (l *Launcher) Launch() error {
	local, err := l.localBuilds()
	if err != nil {
		return err
	}

	var wanted build
	if l.build != 0 {
		builds, err := l.parseBuilds()
		if err != nil {
			return err
		}
		for _, b := range builds {
			if b.buildNumber == l.build {
				wanted = b
				break
			}
		}
		if wanted.uri == "" {
			return fmt.Errorf(
				"Could not find the build you requested, #%d, in the list of available builds", l.build)
		}
	} else {
		var err error
		wanted, err = l.latestBuild()
		if err != nil {
			return err
		}
	}

	if _, exists := local[wanted.buildNumber]; !exists {
		err := l.downloadBuild(wanted)
		if err != nil {
			return err
		}
	}

	return l.launchGame(wanted.buildNumber)
}

func (l *Launcher) latestBuild() (build, error) {
	builds, err := l.parseBuilds()
	if err != nil {
		return build{}, err
	}

	util.Say(l.stdout, "Found %d builds", len(builds))

	localLatest, err := l.latestLocalBuild()
	if err != nil {
		return build{}, err
	}

	if localLatest == 0 {
		util.Say(l.stdout, "No builds have been downloaded yet")
	} else if localLatest != builds[0].buildNumber {
		util.Say(l.stdout, "Latest local build is #%d", localLatest)
		util.Say(
			l.stdout,
			"The latest build is build #%d, released %s",
			builds[0].buildNumber, builds[0].date.Format("2006-01-02 15:04"),
		)
	} else {
		util.Say(
			l.stdout,
			"You have the latest build, #%d, released %s",
			builds[0].buildNumber, builds[0].date.Format("2006-01-02 15:04"),
		)
	}

	return builds[0], nil
}

var fileRE = regexp.MustCompile(`^cataclysmdda-([0-9].[A-Z]-(\d+))\.tar\.gz$`)

func (l *Launcher) parseBuilds() ([]build, error) {
	util.Say(l.stdout, "Getting list of builds from %s", l.buildsURI)
	res, err := http.Get(l.buildsURI)
	if err != nil {
		return []build{}, fmt.Errorf("Could not fetch build list from %s: %s", l.buildsURI, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return []build{}, fmt.Errorf(
			"Did not get a 200 status when fetching %s, got a %d (%s) instead",
			l.buildsURI, res.StatusCode, res.Status,
		)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []build{}, fmt.Errorf("Error parsing HTML from %s: %s", l.buildsURI, err)
	}

	buildDates, err := l.parseBuildDates(doc)
	if err != nil {
		return []build{}, err
	}

	builds := []build{}
	var eachErr error
	doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
		if eachErr != nil {
			return
		}

		href, _ := sel.Attr("href")
		m := fileRE.FindStringSubmatch(href)
		if len(m) < 2 {
			return
		}

		num, err := strconv.Atoi(m[2])
		if err != nil {
			eachErr = fmt.Errorf("Could not convert %s to an integer: %s", m[2], err)
		}

		builds = append(
			builds,
			build{
				uri:         l.buildsURI + href,
				filename:    href,
				version:     m[1],
				buildNumber: uint(num),
				date:        buildDates[href],
			},
		)
	})
	if eachErr != nil {
		return []build{}, eachErr
	}

	// Using After gives us a reverse sorting from most to least recent.
	sort.SliceStable(builds, func(i, j int) bool { return builds[i].date.After(builds[j].date) })
	return builds, nil
}

var buildDatesRE = regexp.MustCompile(`(cataclysmdda-\S+\.tar\.gz)\s+(2\d\d\d-\d\d-\d\d \d\d:\d\d)`)

func (l *Launcher) parseBuildDates(doc *goquery.Document) (map[string]time.Time, error) {
	dates := map[string]time.Time{}
	m := buildDatesRE.FindAllStringSubmatch(doc.Find("body").First().Text(), -1)
	for _, pair := range m {
		d, err := time.Parse("2006-01-02 15:04", pair[2])
		if err != nil {
			return dates, fmt.Errorf("Could not parse date for the file %s from text (%s)", pair[1], pair[2])
		}
		dates[pair[1]] = d
	}
	return dates, nil
}

var buildNumberRE = regexp.MustCompile(`^[1-9][0-9]*$`)

func (l *Launcher) latestLocalBuild() (uint, error) {
	local, err := l.localBuilds()
	if err != nil {
		return 0, err
	}

	nums := []uint{}
	for n := range local {
		nums = append(nums, n)
	}

	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
	return nums[len(nums)-1], nil
}

func (l *Launcher) localBuilds() (map[uint]bool, error) {
	local := map[uint]bool{}

	files, err := ioutil.ReadDir(l.buildDir())
	if err != nil {
		if os.IsNotExist(err) {
			return local, nil
		}
		return local, fmt.Errorf("Could not read directory at %s: %s", l.buildDir(), err)
	}

	for _, f := range files {
		if f.IsDir() && buildNumberRE.MatchString(f.Name()) {
			i, err := strconv.Atoi(f.Name())
			if err != nil {
				return local, fmt.Errorf("Could not convert %s to an integer: %s", f.Name(), err)
			}
			local[uint(i)] = true
		}
	}

	return local, nil
}

func (l *Launcher) downloadBuild(b build) error {
	util.Say(l.stdout, "Downloading build #%d from %s", b.buildNumber, b.uri)

	dir, err := ioutil.TempDir("", "catalauncher-")
	if err != nil {
		return fmt.Errorf("Could not create a temporary directory: %s", err)
	}

	file := filepath.Join(dir, b.filename)
	out, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("Could not create file at %s: %s", file, err)
	}
	defer out.Close()

	resp, err := http.Get(b.uri)
	if err != nil {
		return fmt.Errorf("Could not get %s: %s", b.uri, err)
	}
	defer resp.Body.Close()

	cl := resp.Header.Get("Content-Length")
	len := 0
	if cl != "" {
		len, err = strconv.Atoi(cl)
		if err != nil {
			return fmt.Errorf("Could not convert %s to an integer: %s", cl, err)
		}
	}

	bar := pb.New(len)
	bar.Start()
	rd := bar.NewProxyReader(resp.Body)

	_, err = io.Copy(out, rd)
	if err != nil {
		return fmt.Errorf("Could not save %s to %s: %s", b.uri, file, err)
	}

	return l.untarBuild(file, b)
}

func (l *Launcher) untarBuild(file string, b build) error {
	target := filepath.Join(l.buildDir(), fmt.Sprintf("%d", b.buildNumber))
	err := l.mkdir(target)
	if err != nil {
		return err
	}

	cmd := exec.Command("tar", "xzf", file, "-C", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(`Could not run "tar xvzf %s -C %s": %s\n%s`, file, target, err, out)
	}

	return nil
}

func (l *Launcher) launchGame(buildNumber uint) error {
	dataDir := l.gameDataDir()
	err := l.mkdir(dataDir)
	if err != nil {
		return err
	}

	// XXX - need to get "cataclysmdda-0.C" dynamically
	gameDir := filepath.Join(l.buildDir(), fmt.Sprintf("%d", buildNumber), "cataclysmdda-0.C")

	runPulse := fmt.Sprintf("/run/user/%s/pulse", l.user.Uid)

	args := []string{
		"run",
		// We don't want the container sticking around once the game exits.
		"--rm",
		// We want to make sure save files and such are owned by the current
		// user, not root.
		"--user", fmt.Sprintf("%s:%s", l.user.Uid, l.user.Gid),
		// Needed for sound w/ Pulseaudio
		"-v", "/etc/machine-id:/etc/machine-id",
		"-v", runPulse + ":" + runPulse,
		"-v", "/var/lib/dbus:/var/lib/dbus",
		"-v", fmt.Sprintf("%s/.pulse:/.pulse", l.user.HomeDir),
		// Needed for graphics
		"-e", "DISPLAY",
		"--device", "/dev/dri",
		"-v", "/tmp/.X11-unix:/tmp/.X11-unix",
		//
		"-v", dataDir + ":/data",
		"-v", gameDir + ":/game",
		// CDDA seems to expect PWD to be the game root dir.
		"-w", "/game",
		"houseabsolute/catalauncher-player:latest",
		"./cataclysm-tiles",
		"--savedir", "/data/save/",
		"--configdir", "/data/config/",
		"--memorialdir", "/data/graveyard/",
	}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf("Could not run \"docker %s\": %s\n%s", strings.Join(args, " "), err)
		if len(out) > 0 {
			msg += "\n" + string(out)
		}
		return errors.New(msg)

	}
	os.Exit(0)

	// We should never get here for obvious reasons
	return nil
}

func (l *Launcher) mkdir(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("Could not make directory %s: %s", dir, err)
	}
	return nil
}

func (l *Launcher) gameDataDir() string {
	return filepath.Join(l.config.RootDir(), "game-data")
}

func (l *Launcher) buildDir() string {
	return filepath.Join(l.config.RootDir(), "builds")
}
