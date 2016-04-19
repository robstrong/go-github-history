package main

import (
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
)

type Releases []Release
type ReleaseMap map[string]*Release
type Release struct {
	Version       semver.Version
	PullRequests  []github.PullRequest
	GithubRelease github.RepositoryRelease
}

func (r Release) PublishedDateFormatted() string {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return ""
	}
	if r.GithubRelease.PublishedAt == nil {
		return ""
	}
	return r.GithubRelease.PublishedAt.In(loc).Format("Jan 2, 2006 3:04:05pm")
}

func (r ReleaseMap) MatchToGithubReleases(ghReleases <-chan github.RepositoryRelease) {
	for ghr := range ghReleases {
		tag := strings.TrimLeft(*ghr.TagName, "v")
		if rel, ok := r[tag]; ok {
			rel.GithubRelease = ghr
			debugLog("Found Github release for tag %s: %s", tag, *ghr.Name)
		} else {
			debugLog("Could not find Github release for tag %s", tag)
		}
	}
}
