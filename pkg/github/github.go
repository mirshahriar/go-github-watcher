package github

import (
	"context"
	"sync"

	"github.com/aerokite/go-github-watcher/pkg/transport"
	"github.com/google/go-github/github"
)

func NewGithubClient() *github.Client {
	return github.NewClient(transport.NewMemoryCacheTransport().Client())
}

func NewGithubClientWithToken(token string) *github.Client {
	return github.NewClient(transport.NewMemoryCacheTransport().SetToken(token).Client())
}

func NewBiblio(client *github.Client) *Biblio {
	biblio := &Biblio{
		Cache:  make(map[string]*RepositoryInfo),
		Client: client,
	}
	if client == nil {
		biblio.Client = NewGithubClient()
	}

	return biblio
}

func (b *Biblio) listRepositoriesByOrg(org string) ([]*github.Repository, error) {
	allRepositories := make([]*github.Repository, 0)
	for i := 1; ; i++ {
		repositories, _, err := b.Client.Repositories.ListByOrg(context.Background(), org,
			&github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{
					Page:    i,
					PerPage: 100,
				},
			},
		)
		if err != nil {
			return nil, err
		}
		allRepositories = append(allRepositories, repositories...)
	}
	return allRepositories, nil
}

func (b *Biblio) getRepositories(org string, repositories ...string) ([]*github.Repository, error) {
	var err error
	var allRepositories []*github.Repository
	if len(repositories) == 0 {
		allRepositories, err = b.listRepositoriesByOrg(org)
		if err != nil {
			return nil, err
		}
	} else {
		for _, repo := range repositories {
			repository, _, err := b.Client.Repositories.Get(context.Background(), org, repo)
			if err != nil {
				return nil, err
			}

			allRepositories = append(allRepositories, repository)
		}
	}
	return allRepositories, nil
}

func (b *Biblio) countNewOpenIssues(org, repo string, lastSyncedIssue int) (int, int, error) {
	newLastSyncedIssue := lastSyncedIssue
	count := 0
	var once sync.Once

	for i := 1; ; i++ {
		var issues []*github.Issue
		issues, _, err := b.Client.Issues.ListByRepo(context.Background(), org, repo, &github.IssueListByRepoOptions{
			ListOptions: github.ListOptions{
				Page:    i,
				PerPage: 100,
			},
		})
		if err != nil {
			return 0, 0, err
		}
		if len(issues) == 0 {
			return 0, newLastSyncedIssue, nil
		}

		for _, issue := range issues {
			if *issue.Number <= lastSyncedIssue {
				return count, newLastSyncedIssue, nil
			} else {
				once.Do(func() {
					newLastSyncedIssue = *issue.Number
				})
				count++
			}
		}
	}
	return count, newLastSyncedIssue, nil
}

func (b *Biblio) getStargazers(org, repo string) ([]string, error) {
	users := make([]string, 0)
	for i := 1; ; i++ {
		stargazers, _, err := b.Client.Activity.ListStargazers(context.Background(), org, repo,
			&github.ListOptions{
				Page:    i,
				PerPage: 100,
			},
		)
		if err != nil {
			return nil, err
		}
		if len(stargazers) == 0 {
			return users, nil
		}

		for _, stargazer := range stargazers {
			users = append(users, *(stargazer.User.Login))
		}
	}
	return users, nil
}

func (b *Biblio) GetRepositoriesInfo(org string, repositoris ...string) (map[string]*RepositoryInfo, error) {
	allRepositories, err := b.getRepositories(org, repositoris...)
	if err != nil {
		return nil, err
	}

	cachedOrganizationReposInfo := b.Cache
	newOrganizationReposInfoMap := make(map[string]*RepositoryInfo)

	for _, repo := range allRepositories {
		repoName := ""
		if repo.Name != nil {
			repoName = *repo.Name
		}

		if repoName == "" {
			continue
		}

		cachedRepoInfo := cachedOrganizationReposInfo[repoName]
		repoInfo := new(RepositoryInfo)

		// Track Issues
		lastSyncedIssue := 0
		if cachedRepoInfo != nil {
			lastSyncedIssue = cachedRepoInfo.LastSyncedIssue.IssueNumber
		}
		count, issueNumber, err := b.countNewOpenIssues(org, repoName, lastSyncedIssue)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			repoInfo.LastSyncedIssue.IssueNumber = issueNumber
			repoInfo.LastSyncedIssue.Count = count
		}

		// Track Stargazers
		users, err := b.getStargazers(org, repoName)
		if err != nil {
			return nil, err
		}
		repoInfo.Stargazers = users

		// Track
		if repo.ForksCount != nil {
			repoInfo.ForksCount = *repo.ForksCount
		}

		newOrganizationReposInfoMap[repoName] = repoInfo
	}
	return newOrganizationReposInfoMap, nil
}

func (b *Biblio) InitializeCache(org string, repositories ...string) error {
	repositoriesInfo, err := b.GetRepositoriesInfo(org, repositories...)
	if err != nil {
		return err
	}
	b.Cache = repositoriesInfo
	return nil
}
