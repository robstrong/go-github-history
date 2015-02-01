package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os/user"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	githubRepo   = kingpin.Arg("repo", "Github Repository in the format 'owner/repository'").Required().String()
	tokenPath    = kingpin.Flag("token-path", "Path to file containing token").Short('t').String()
	outputFile   = kingpin.Flag("out", "HTML output file").Short('o').Default("gh-history.html").String()
	templateFile = kingpin.Flag("template", "HTML template file").Default("template.html").String()
	debug        = kingpin.Flag("verbose", "Enable verbose output").Default("false").Bool()
)

func main() {
	kingpin.Parse()

	history := ReleaseHistory{
		GenTime:    time.Now().Format("Jan 2, 2006 3:04:05pm"),
		Repository: *githubRepo,
	}

	//get token from home dir
	token, err := getToken(*tokenPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	//setup git repo (used to lookup tags for PRs)
	log.Printf("Setting up repository")
	repo, err := NewRepo(*githubRepo)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = repo.SetupRepo(token)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Repository ready")

	log.Printf("Gathering pull requests")
	//setup github client
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: token},
	}
	client := github.NewClient(t.Client())
	prs := getAllPullRequests(client, repo.Owner, repo.Repository)
	releases := getAllReleases(client, repo.Owner, repo.Repository)

	//look at the head SHA for each PR and find the earliest git tag that contains
	//that commit. that tag is the version that the PR was merged into
	for pr := range prs {
		sha := *pr.Head.SHA
		ver, err := repo.getVersionForCommit(sha)
		if err != nil {
			debugLog("ERROR: %s in PR #%d", err.Error(), *pr.Number)
			continue
		}
		history.AddPullRequest(ver, pr)
	}
	log.Printf("Pull requests collected")

	history.Releases.MatchToGithubReleases(releases)

	history.GenerateHTML(*outputFile)
}

func debugLog(format string, a ...interface{}) {
	if *debug {
		log.Printf(format, a...)
	}
}

func getToken(tokenPath string) (token string, err error) {
	//check if a token file path was passed in
	if tokenPath != "" {
		contents, readErr := ioutil.ReadFile(tokenPath)
		if readErr != nil {
			err = errors.New("Could not find token file at path: " + tokenPath)
			return
		}
		token = strings.TrimSpace(string(contents))
		return
	}

	//check if .gh-history file exists in home dir
	usr, err := user.Current()
	if err != nil {
		return
	}
	homeToken, err := ioutil.ReadFile(usr.HomeDir + "/.gh-history/token")
	if err != nil {
		err = errors.New("Github Token not found. You must create a file in " +
			usr.HomeDir + "/.gh-history/token which contains your token. To create a " +
			"token go to, https://github.com/settings/applications#personal-access-tokens")
		return
	}
	token = strings.TrimSpace(string(homeToken))
	return
}
