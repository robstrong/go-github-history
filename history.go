package main

import (
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
)

type ReleaseHistory struct {
	Repository string
	Releases   ReleaseMap
	GenTime    string
}

func (h ReleaseHistory) SortedReleases() Releases {
	var versions semver.Versions
	for v := range h.Releases {
		ver, err := semver.New(v)
		if err != nil {
			debugLog("Invalid version: " + err.Error())
			continue
		}
		versions = append(versions, ver)
	}
	semver.Sort(versions)
	var sorted Releases
	for i := len(versions) - 1; i >= 0; i-- {
		sorted = append(sorted, *h.Releases[versions[i].String()])
	}
	return sorted
}

func (h *ReleaseHistory) AddPullRequest(v semver.Version, pr github.PullRequest) {
	if h.Releases == nil {
		h.Releases = make(map[string]*Release)
	}
	if _, ok := h.Releases[v.String()]; !ok {
		h.Releases[v.String()] = &Release{
			Version: v,
		}
	}
	h.Releases[v.String()].PullRequests = append(h.Releases[v.String()].PullRequests, pr)
}

func (h *ReleaseHistory) GenerateHTML(o string) {
	log.Print("Generating HTML")
	tmpl, err := template.ParseFiles(*templateFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	out, err := os.Create(o)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer out.Close()
	err = tmpl.Execute(out, *h)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Print("Done Generating HTML")
}

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
