package util

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func RepoFromRef(ref string) (url string, branch string) {
	// given reference, for example, https://github.com/user/repo@branch-or-tag
	// return repo url: https://github.com/user/repo, branch: branch-or-tag
	githubPattern := regexp.MustCompile(`(?P<url>https?://[\w.-]+/[\w.-]+/[\w.-]+)(@(?P<branch>[^@]+))?$`)
	match := githubPattern.FindStringSubmatch(ref)
	if match == nil {
		return
	}
	m := map[string]string{}
	for i, n := range githubPattern.SubexpNames() {
		if n != "" {
			m[n] = match[i]
		}
	}
	url = m["url"]
	branch = m["branch"]
	return
}

// CloneGitRepo(d.Repo, d.getRepoLocalPath())
func CloneGitRepo(ref string, path string) {
	// split repo url from ref
	log.Printf("Checking out repo %s...", ref)
	url, branch := RepoFromRef(ref)
	if url == "" {
		log.Panicf("Invalid repo %s", ref)
	}
	log.Printf("Check if local repo exist by: %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Clonning the repo %s...", url)
		gitClone := fmt.Sprintf("git clone %s %s", url, path)
		if branch != "" {
			gitClone = fmt.Sprintf("git clone %s -b %s %s", url, branch, path)
		}
		Shell(gitClone)

	} else {
		log.Printf("Local git repo for %s already exists, skipping...", url)
	}
}

func GetRepoHead(localPath string) string {
	commit, _ := ShellOutput(
		fmt.Sprintf("cd %v & git rev-parse --short HEAD", localPath),
	)
	return strings.TrimSpace(commit)
}
