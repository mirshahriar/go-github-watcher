package github

import (
	"github.com/google/go-github/github"
)

type Biblio struct {
	Cache  map[string]*RepositoryInfo
	Client *github.Client
}

type RepositoryInfo struct {
	Stargazers      []string
	Subscribers     []string
	LastSyncedIssue struct {
		IssueNumber int
		Count       int
	}
	LastPR        int
	ReleasesCount int
	ForksCount    int
}
