package main

import (
	"html/template"
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
)

type History struct {
	Repository string
	Releases   ReleaseMap
	Issues     []github.Issue
	GenTime    string
}

func (h *History) GenerateHTML(o string) {
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

func (h History) SortedReleases() Releases {
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

func (h *History) AddPullRequest(v semver.Version, pr github.PullRequest) {
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
