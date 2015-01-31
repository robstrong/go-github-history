package main

import (
	"log"
	"os"
	"text/template"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
)

type ReleaseHistory struct {
	Releases map[string]*Release
	GenTime  string
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
type Release struct {
	Version      semver.Version
	PullRequests []github.PullRequest
}
