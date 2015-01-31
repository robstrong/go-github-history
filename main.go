package main

import (
	"log"
	"sync"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	githubOwner  = kingpin.Flag("owner", "Github Repository Owner").Short('o').Required().String()
	githubRepo   = kingpin.Flag("repo", "Github Repository").Short('r').Required().String()
	token        = kingpin.Flag("token", "Token").Short('t').Required().String()
	outputFile   = kingpin.Flag("out", "HTML output file").Default("gh-history.html").String()
	templateFile = kingpin.Flag("template", "HTML template file").Default("template.html").String()
	debug        = kingpin.Flag("verbose", "Enable verbose output").Default("false").Bool()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	history := ReleaseHistory{
		GenTime: time.Now().Format("Jan 2, 2006 3:04:05pm"),
	}

	//setup git repo (used to lookup tags for PRs)
	log.Printf("Setting up repository")
	repo := NewRepo(*githubOwner, *githubRepo)
	err := repo.SetupRepo(*token)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Repository ready")

	log.Printf("Gathering pull requests")
	//setup github client
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: *token},
	}
	client := github.NewClient(t.Client())
	prs := getAllPullRequests(client, *githubOwner, *githubRepo)

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

	history.GenerateHTML(*outputFile)
}

func debugLog(format string, a ...interface{}) {
	if *debug {
		log.Printf(format, a...)
	}
}

func getAllPullRequests(client *github.Client, owner string, repo string) <-chan github.PullRequest {
	out := make(chan github.PullRequest)
	go func() {
		opt := &github.PullRequestListOptions{
			State: "closed",
		}
		opt.ListOptions.PerPage = 100
		debugLog("Fetching Page: 1")
		prs, resp, err := client.PullRequests.List(
			owner,
			repo,
			opt,
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		debugLog("Found %d PRs", len(prs))
		for _, pr := range prs {
			out <- pr
		}
		if resp.NextPage == 0 {
			debugLog("All PRs collected")
			close(out)
			return
		}
		var group sync.WaitGroup
		for i := 2; i <= resp.LastPage; i++ {
			group.Add(1)
			go func(page int) {
				debugLog("Fetching Page: %d", page)
				opt.ListOptions.Page = page
				prs, _, err := client.PullRequests.List(
					owner,
					repo,
					opt,
				)
				debugLog("Found %d PRs on page %d", len(prs), page)
				if err != nil {
					log.Fatal(err.Error())
				}
				for _, pr := range prs {
					out <- pr
				}
				group.Done()
			}(i)
		}
		go func() {
			group.Wait()
			close(out)
		}()
	}()
	return out
}
