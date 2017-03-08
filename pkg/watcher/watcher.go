package watcher

import (
	"errors"
	"log"
	"os"

	"github.com/aerokite/go-github-watcher/pkg/github"
	"github.com/robfig/cron"
)

func New() *watcher {
	w := &watcher{
		repositories: make([]string, 0),
		store:        make(map[string]*github.RepoInfo),
	}
	return w
}

func (w *watcher) SetOrganization(name string) {
	w.organization = name
}

func (w *watcher) AddRepositories(names ...string) {
	w.repositories = append(w.repositories, names...)
}

func (w *watcher) Schedule(expr string) (*job, error) {
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

	j.initializeStore()

	if err := j.cron.AddFunc(expr, j.watch); err != nil {
		return nil, err
	}
	return j, nil
}

func (j *job) initializeStore() {
	setData := func() {
		data, err := github.GetData(j.organization, j.repositories, j.store)
		if err != nil {
			log.Fatalln(err)
		}
		j.store = data
	}
	j.Once.Do(setData)
}

func (j *job) watch() {
	newData, err := github.GetData(j.organization, j.repositories, j.store)
	if err != nil {
		log.Println(err.Error())
	} else {
		comapareData(j.store, newData)
		j.store = newData
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
