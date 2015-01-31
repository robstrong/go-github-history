package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
)

var basePath = "repos/"

type Repo struct {
	Owner      string
	Repository string
}

func (r *Repo) getVersionForCommit(sha string) (semver.Version, error) {
	gitPath, err := exec.LookPath("git")
	cmd := exec.Command(gitPath, "tag", "--sort=v:refname", "--contains", sha)
	cmd.Dir = r.GetRepoPath()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return semver.Version{}, errors.New(string(out))
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 1 {
		return semver.Version{}, errors.New("Tag not found")
	}
	ver, err := semver.New(strings.Replace(lines[0], "v", "", 1))
	if err != nil {
		return ver, errors.New(err.Error() + ", tag: " + lines[0])
	}
	return ver, nil
}

func (r *Repo) GetRepoPath() (path string) {
	path = basePath + r.Owner + "/" + r.Repository
	return
}
func (r *Repo) SetupRepo(token string) (err error) {
	gitPath, err := exec.LookPath("git")
	path := r.GetRepoPath()
	//if path doesn't exist, create and init git
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Cloning repo to %s", path)
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return err
		}
		cmd := exec.Command(gitPath, "init")
		cmd.Dir = path
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	cmd := exec.Command(gitPath, "pull", "--tags", "https://"+token+"@github.com/"+r.Owner+"/"+r.Repository+".git")
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		return
	}

	return nil
}

func NewRepo(owner string, repo string) *Repo {
	return &Repo{
		Owner:      owner,
		Repository: repo,
	}
}
