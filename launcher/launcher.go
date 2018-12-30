package launcher

import (
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
	"github.com/houseabsolute/catalauncher/util"
	pb "gopkg.in/cheggaaa/pb.v2"
)

type build struct {
	uri         string
	filename    string
	version     string
	buildNumber int
	date        time.Time
}

type C struct {
	rootDir     string
	stdout      io.Writer
	stderr      io.Writer
	buildsURI   string
	currentUser *user.User
}

const defaultBuildsURI = "http://dev.narc.ro/cataclysm/jenkins-latest/Linux_x64/Tiles/"

func New(rootDir string) *C {
	return &C{
		rootDir:   rootDir,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		buildsURI: defaultBuildsURI,
	}
}

func (c *C) Launch() {
	builds := c.parseBuilds()
	util.Say(c.stdout, "Found %d builds", len(builds))

	cur := c.currentBuild()
	if cur == 0 {
		util.Say(c.stdout, "No builds have been downloaded yet")
	} else if cur != builds[0].buildNumber {
		util.Say(c.stdout, "Currently using build #%d", cur)
		util.Say(
			c.stdout,
			"The latest build is build #%d, released %s",
			builds[0].buildNumber, builds[0].date.Format("2006-01-02 15:04"),
		)
	} else {
		util.Say(
			c.stdout,
			"You have the latest build, #%d, released %s",
			builds[0].buildNumber, builds[0].date.Format("2006-01-02 15:04"),
		)
	}

	if cur != builds[0].buildNumber {
		c.downloadBuild(builds[0])
		cur = builds[0].buildNumber
	}

	c.launchGame(cur)
}

var fileRE = regexp.MustCompile(`^cataclysmdda-([0-9].[A-Z]-(\d+))\.tar\.gz$`)

func (c *C) parseBuilds() []build {
	util.Say(c.stdout, "Getting list of builds from %s", c.buildsURI)
	res, err := http.Get(c.buildsURI)
	if err != nil {
		c.printErrorAndExit("Could not fetch build list from %s: %s", c.buildsURI, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.printErrorAndExit(
			"Did not get a 200 status when fetching %s, got a %d (%s) instead",
			c.buildsURI, res.StatusCode, res.Status,
		)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		c.printErrorAndExit("Error parsing HTML from %s: %s", c.buildsURI, err)
	}

	buildDates := c.parseBuildDates(doc)

	builds := []build{}
	doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		m := fileRE.FindStringSubmatch(href)
		if len(m) < 2 {
			return
		}

		num, err := strconv.Atoi(m[2])
		if err != nil {
			c.printErrorAndExit("Could not convert %s to an integer: %s", m[2], err)
		}

		builds = append(
			builds,
			build{
				uri:         c.buildsURI + href,
				filename:    href,
				version:     m[1],
				buildNumber: num,
				date:        buildDates[href],
			},
		)
	})
	// Using After gives us a reverse sorting from most to least recent.
	sort.SliceStable(builds, func(i, j int) bool { return builds[i].date.After(builds[j].date) })
	return builds
}

var buildDatesRE = regexp.MustCompile(`(cataclysmdda-\S+\.tar\.gz)\s+(2\d\d\d-\d\d-\d\d \d\d:\d\d)`)

func (c *C) parseBuildDates(doc *goquery.Document) map[string]time.Time {
	dates := map[string]time.Time{}
	m := buildDatesRE.FindAllStringSubmatch(doc.Find("body").First().Text(), -1)
	for _, pair := range m {
		d, err := time.Parse("2006-01-02 15:04", pair[2])
		if err != nil {
			c.printErrorAndExit("Could not parse date for the file %s from text (%s)", pair[1], pair[2])
		}
		dates[pair[1]] = d
	}
	return dates
}

var buildNumberRE = regexp.MustCompile(`^[1-9][0-9]*$`)

func (c *C) currentBuild() int {
	files, err := ioutil.ReadDir(c.buildDir())
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		c.printErrorAndExit("Could not read directory at %s: %s", c.buildDir(), err)
	}

	builds := []int{}
	for _, f := range files {
		if f.IsDir() && buildNumberRE.MatchString(f.Name()) {
			i, err := strconv.Atoi(f.Name())
			if err != nil {
				c.printErrorAndExit("Could not convert %s to an integer: %s", f.Name(), err)
			}
			builds = append(builds, i)
		}
	}

	sort.Ints(builds)

	return builds[len(builds)-1]
}

func (c *C) downloadBuild(b build) {
	util.Say(c.stdout, "Downloading build #%d from %s", b.buildNumber, b.uri)

	dir, err := ioutil.TempDir("", "catalauncher-")
	if err != nil {
		c.printErrorAndExit("Could not create a temporary directory: %s", err)
	}

	file := filepath.Join(dir, b.filename)
	out, err := os.Create(file)
	if err != nil {
		c.printErrorAndExit("Could not create file at %s: %s", file, err)
	}
	defer out.Close()

	resp, err := http.Get(b.uri)
	if err != nil {
		c.printErrorAndExit("Could not get %s: %s", b.uri, err)
	}
	defer resp.Body.Close()

	cl := resp.Header.Get("Content-Length")
	len := 0
	if cl != "" {
		len, err = strconv.Atoi(cl)
		if err != nil {
			c.printErrorAndExit("Could not convert %s to an integer: %s", cl, err)
		}
	}

	bar := pb.New(len)
	bar.Start()
	rd := bar.NewProxyReader(resp.Body)

	_, err = io.Copy(out, rd)
	if err != nil {
		c.printErrorAndExit("Could not save %s to %s: %s", b.uri, file, err)
	}

	c.untarBuild(file, b)
}

func (c *C) untarBuild(file string, b build) string {
	target := filepath.Join(c.buildDir(), strconv.Itoa(b.buildNumber))
	c.mkdir(target)

	cmd := exec.Command("tar", "xzf", file, "-C", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.printErrorAndExit(`Could not run "tar xvzf %s -C %s": %s\n%s`, file, target, err, out)
	}

	return ""
}

func (c *C) launchGame(num int) {
	dataDir := c.gameDataDir()
	c.mkdir(dataDir)

	// XXX - need to get "cataclysmdda-0.C" dynamically
	gameDir := filepath.Join(c.buildDir(), strconv.Itoa(num), "cataclysmdda-0.C")

	args := []string{
		"run",
		"--user", fmt.Sprintf("%s:%s", c.userID(), c.groupID()),
		"--rm",
		"-i",
		"-e", "DISPLAY",
		"--device", "/dev/dri",
		"--device", "/dev/snd",
		"-v", "/tmp/.X11-unix:/tmp/.X11-unix",
		"-v", dataDir + ":/data",
		"-v", gameDir + ":/game",
		"-w", "/game",
		"houseabsolute/catalauncher-player:latest",
		"./cataclysm-tiles", "--save-dir", "/data/save", "--config-dir", "/data/config",
	}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.printErrorAndExit("Could not run \"docker %s\": %s\n%s", strings.Join(args, " "), err, out)

	}
	os.Exit(0)
}

func (c *C) mkdir(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		c.printErrorAndExit("Could not make directory %s: %s", dir, err)
	}
}

func (c *C) userID() string {
	return c.user().Uid
}

func (c *C) groupID() string {
	return c.user().Gid
}

func (c *C) user() *user.User {
	if c.currentUser != nil {
		return c.currentUser
	}
	u, err := user.Current()
	if err != nil {
		c.printErrorAndExit("Could not get the current user: %s", err)
	}
	c.currentUser = u
	return c.currentUser
}

func (c *C) gameDataDir() string {
	return filepath.Join(c.rootDir, "game-data")
}

func (c *C) buildDir() string {
	return filepath.Join(c.rootDir, "builds")
}

func (c *C) printErrorAndExit(tmpl string, args ...interface{}) {
	util.Say(c.stderr, tmpl, args...)
	os.Exit(1)
}
