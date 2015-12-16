package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	genType      = kingpin.Arg("gen-type", "Generation type ('releases' or 'issues')").Required().String()
	githubRepo   = kingpin.Arg("repo", "Github Repository in the format 'owner/repository'").Required().String()
	tokenPath    = kingpin.Flag("token-path", "Path to file containing token").Short('t').String()
	outputFile   = kingpin.Flag("out", "HTML output file").Short('o').Default("gh-history.html").String()
	templateFile = kingpin.Flag("template", "HTML template file, default is 'releases.html' or 'issues.html' depending on gen-type").Default("").String()
	debug        = kingpin.Flag("verbose", "Enable verbose output").Default("false").Bool()
)

func main() {
	kingpin.Parse()

	history := History{
		GenTime:    time.Now().Format("Jan 2, 2006 3:04:05pm"),
		Repository: *githubRepo,
	}

	//get token from home dir
	token, err := getToken(*tokenPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	//setup github client
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: token},
	}
	client := github.NewClient(t.Client())
	repo, err := NewRepo(*githubRepo)
	currentDir := curDir()

	switch *genType {

	case "releases":
		//setup git repo (used to lookup tags for PRs)
		log.Printf("Setting up repository")
		if err != nil {
			log.Fatal(err.Error())
		}
		err = repo.SetupRepo(token)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Repository ready")

		log.Printf("Gathering pull requests and releases")
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
		log.Printf("Pull requests and releases collected")

		history.Releases.MatchToGithubReleases(releases)
		if *templateFile == "" {
			*templateFile = fmt.Sprintf("%s/%s", currentDir, "releases.html")
		}

	case "issues":
		closedIssues := getAllIssues(client, repo.Owner, repo.Repository, "closed")
		for issue := range closedIssues {
			if issue.PullRequestLinks == nil {
				history.Issues = append(history.Issues, issue)
			}
		}
		openIssues := getAllIssues(client, repo.Owner, repo.Repository, "open")
		for issue := range openIssues {
			if issue.PullRequestLinks == nil {
				history.Issues = append(history.Issues, issue)
			}
		}
		sort.Sort(ByNumber(history.Issues))
		if *templateFile == "" {
			*templateFile = fmt.Sprintf("%s/%s", currentDir, "issues.html")
		}

	default:
		log.Printf("Invalid generation type: %s\n", *genType)
	}
	history.GenerateHTML(*outputFile)
}

func debugLog(format string, a ...interface{}) {
	if *debug {
		log.Printf(format, a...)
	}
}

func curDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
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
