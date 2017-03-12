package watcher

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aerokite/go-github-watcher/pkg/github"
	gogithub "github.com/google/go-github/github"
	"github.com/robfig/cron"
)

func New() *watcher {
	w := &watcher{
		repositories: make([]string, 0),
	}
	return w
}

func (w *watcher) SetGithubToken(token string) {
	w.githubToken = token
}

func (w *watcher) SetOrganization(name string) {
	w.organization = name
}

func (w *watcher) AddRepositories(names ...string) {
	w.repositories = append(w.repositories, names...)
}

func (w *watcher) Schedule(expr string) (*job, error) {
	if w.githubToken == "" {
		w.biblio = github.NewBiblio(nil)
	} else {
		w.biblio = github.NewBiblio(github.NewGithubClientWithToken(w.githubToken))
	}

	j := &job{
		watcher: w,
		cron:    cron.New(),
	}
	if j.organization == "" {
		return nil, errors.New(`Must provide organization name`)
	}
	if expr == "" {
		return nil, errors.New(`Must provide cron expression`)
	}

	j.initializeCache()

	if err := j.cron.AddFunc(expr, j.watch); err != nil {
		return nil, err
	}
	return j, nil
}

func (j *job) initializeCache() {
	setData := func() {
		j.biblio.InitializeCache(j.organization, j.repositories...)
	}
	j.Once.Do(setData)
}

func (j *job) watch() {
	organizationReposInfoMap, err := j.biblio.GetRepositoriesInfo(j.organization, j.repositories...)
	if err != nil {
		if rateLimitError, ok := err.(*gogithub.RateLimitError); ok {
			log.Println("Hit rate limit.",
				fmt.Sprintf("Rate reset in %v at %v",
					rateLimitError.Rate.Reset.Time.Sub(time.Now()),
					rateLimitError.Rate.Reset.UTC().String()))
		}
	} else {
		comapareData(j.biblio.Cache, organizationReposInfoMap)
		j.biblio.Cache = organizationReposInfoMap
	}
}

func (j *job) RunAndHold() {
	j.Run()
	sigs := make(chan os.Signal, 1)
	<-sigs
}

func (j *job) Run() {
	j.cron.Start()
}

func (j *job) Stop() {
	j.cron.Stop()
}
