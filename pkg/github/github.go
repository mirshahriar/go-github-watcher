package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	//"strings"
	"log"
	"strings"
	"sync"
	"time"
)

func CheckRateLimit() (bool, error) {
	rateLimitUrl := "https://api.github.com/rate_limit"
	response, err := http.Get(rateLimitUrl)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	var rateLimit RateLimit
	if err := json.NewDecoder(response.Body).Decode(&rateLimit); err != nil {
		return false, err
	}

	if rateLimit.Rate.Remaining == 0 {
		now := time.Now()
		resetTime := time.Unix(int64(rateLimit.Rate.Reset), 0)
		log.Println(
			fmt.Sprintf("API rate limit exceeded. Rate limit will be reset after %d minute(s)",
				int(resetTime.Sub(now).Minutes())),
		)
		return false, nil
	}
	return true, nil
}

func GetAllRepos(org string) ([]*GithubRepo, error) {
	orgRepoUrl := fmt.Sprintf("https://api.github.com/orgs/%v/repos", org)

	allGithubRepos := make([]*GithubRepo, 0)
	for i := 1; ; i++ {
		response, err := http.Get(orgRepoUrl + fmt.Sprintf("?page=%v", i))
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		var repos []*GithubRepo
		if err := json.NewDecoder(response.Body).Decode(&repos); err != nil {
			return nil, err
		}

		if len(repos) == 0 {
			break
		}
		allGithubRepos = append(allGithubRepos, repos...)
	}

	return allGithubRepos, nil
}

func GetRepo(orgRepoUrl string) (*GithubRepo, error) {
	response, err := http.Get(orgRepoUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var repo GithubRepo
	if err := json.NewDecoder(response.Body).Decode(&repo); err != nil {
		return nil, err
	}

	return &repo, nil
}

func CountNewOpenIssues(issuesUrl string, lastSyncedOpenIssue int) (count int64, lastOpenIssue int, err error) {
	var once sync.Once
	issuesUrl = strings.TrimRight(issuesUrl, "{/number}")
	lastOpenIssue = lastSyncedOpenIssue
	var response *http.Response
	for i := 1; ; i++ {
		response, err = http.Get(issuesUrl + fmt.Sprintf("?page=%v", i))
		if err != nil {
			return
		}
		defer response.Body.Close()

		var issues RepoIssues
		if err = json.NewDecoder(response.Body).Decode(&issues); err != nil {
			return
		}

		if len(issues) == 0 {
			return
		}

		for _, issue := range issues {
			if issue.Number <= lastSyncedOpenIssue {
				return
			} else {
				once.Do(func() {
					lastOpenIssue = issue.Number
				})
				count++
			}
		}
	}

	return
}

func GetStargazers(stargazersURL string) (users []string, err error) {
	var response *http.Response
	for i := 1; ; i++ {
		response, err = http.Get(stargazersURL + fmt.Sprintf("?page=%v", i))
		if err != nil {
			return
		}
		defer response.Body.Close()

		var stargazers RepoStargazers
		if err = json.NewDecoder(response.Body).Decode(&stargazers); err != nil {
			return
		}

		if len(stargazers) == 0 {
			return
		}

		for _, user := range stargazers {
			users = append(users, user.Login)
		}
	}
	return
}

func GetData(org string, repositoris []string, storedData map[string]*RepoInfo) (map[string]*RepoInfo, error) {
	newData := make(map[string]*RepoInfo)

	letMeCheck, err := CheckRateLimit()
	if err != nil {
		return nil, err
	}
	if !letMeCheck {
		return storedData, nil
	}

	var allGithubRepos []*GithubRepo
	if len(repositoris) == 0 {
		allGithubRepos, err = GetAllRepos(org)
		if err != nil {
			return nil, err
		}
	} else {
		for _, repo := range repositoris {
			url := fmt.Sprintf("https://api.github.com/repos/%v/%v", org, repo)
			gitRepo, err := GetRepo(url)
			if err != nil {
				return nil, err
			}
			allGithubRepos = append(allGithubRepos, gitRepo)
		}
	}

	for _, repo := range allGithubRepos {
		repoInfo := new(RepoInfo)

		// Track Issues
		lastSyncedIssue := 0
		if info, found := storedData[repo.Name]; found {
			lastSyncedIssue = info.LastSyncedIssue.IssueNumber
		}
		count, number, err := CountNewOpenIssues(repo.IssuesURL, lastSyncedIssue)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			repoInfo.LastSyncedIssue.IssueNumber = number
			repoInfo.LastSyncedIssue.Count = count
		}

		// Track Stargazers
		users, err := GetStargazers(repo.StargazersURL)
		if err != nil {
			return nil, err
		}
		repoInfo.Stargazers = users

		newData[repo.Name] = repoInfo
	}

	return newData, nil
}
