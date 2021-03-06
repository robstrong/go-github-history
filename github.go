package main

import (
	"log"
	"sync"

	"github.com/google/go-github/github"
)

func getAllReleases(client *github.Client, owner string, repo string) <-chan github.RepositoryRelease {
	out := make(chan github.RepositoryRelease)
	go func() {
		opt := &github.ListOptions{
			PerPage: 100,
		}
		debugLog("Fetching Page: 1")
		rels, resp, err := client.Repositories.ListReleases(
			owner,
			repo,
			opt,
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		debugLog("Found %d Releases", len(rels))
		for _, rel := range rels {
			out <- rel
		}
		if resp.NextPage == 0 {
			debugLog("All releases collected")
			close(out)
			return
		}
		var group sync.WaitGroup
		for i := 2; i <= resp.LastPage; i++ {
			group.Add(1)
			go func(page int) {
				debugLog("Fetching Release Page: %d", page)
				opt.Page = page
				rels, _, err := client.Repositories.ListReleases(
					owner,
					repo,
					opt,
				)
				debugLog("Found %d releases on page %d", len(rels), page)
				if err != nil {
					log.Fatal(err.Error())
				}
				for _, rel := range rels {
					out <- rel
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

func getAllPullRequests(client *github.Client, owner string, repo string) <-chan github.PullRequest {
	out := make(chan github.PullRequest)
	go func() {
		opt := &github.PullRequestListOptions{
			State: "closed",
		}
		opt.ListOptions.PerPage = 100
		debugLog("Fetching PR Page: 1")
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
				debugLog("Fetching PR Page: %d", page)
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

func getAllIssues(client *github.Client, owner string, repo string, state string) <-chan github.Issue {
	out := make(chan github.Issue)
	go func() {
		opt := &github.IssueListByRepoOptions{
			State:     state,
			Sort:      "created",
			Direction: "desc",
		}
		opt.ListOptions.PerPage = 100
		debugLog("Fetching Issue Page: 1")
		issues, resp, err := client.Issues.ListByRepo(
			owner,
			repo,
			opt,
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		debugLog("Found %d Issues", len(issues))
		for _, issue := range issues {
			out <- issue
		}
		if resp.NextPage == 0 {
			debugLog("All Issues Collected")
			close(out)
			return
		}
		var group sync.WaitGroup
		for i := 2; i <= resp.LastPage; i++ {
			group.Add(1)
			go func(page int) {
				debugLog("Fetching Issue Page: %d", page)
				opt.ListOptions.Page = page
				issues, _, err := client.Issues.ListByRepo(
					owner,
					repo,
					opt,
				)
				debugLog("Found %d Issues on page %d", len(issues), page)
				if err != nil {
					log.Fatal(err.Error())
				}
				for _, issue := range issues {
					out <- issue
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

type ByNumber []github.Issue

func (a ByNumber) Len() int {
	return len(a)
}
func (a ByNumber) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByNumber) Less(i, j int) bool {
	return *a[i].Number > *a[j].Number
}
