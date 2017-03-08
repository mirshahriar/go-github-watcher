package github

type RateLimit struct {
	Rate struct {
		Limit     int `json:"limit"`
		Remaining int `json:"remaining"`
		Reset     int `json:"reset"`
	} `json:"rate"`
}

type GithubRepo struct {
	Name           string `json:"name"`
	FullName       string `json:"full_name"`
	StargazersURL  string `json:"stargazers_url"`
	SubscribersURL string `json:"subscribers_url"`
	IssuesURL      string `json:"issues_url"`
	PullsURL       string `json:"pulls_url"`
	ReleasesURL    string `json:"releases_url"`
	ForksCount     int    `json:"forks_count"`
}

type RepoInfo struct {
	Stargazers      []string
	Subscribers     []string
	LastSyncedIssue struct {
		IssueNumber int
		Count       int64
	}
	LastPR        int64
	ReleasesCount int64
	ForksCount    int64
}

type RepoIssues []struct {
	Number int `json:"number"`
}

type RepoStargazers []struct {
	Login string `json:"login"`
}
