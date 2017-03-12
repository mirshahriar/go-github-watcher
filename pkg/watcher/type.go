package watcher

import (
	"sync"

	"github.com/aerokite/go-github-watcher/pkg/github"
	"github.com/robfig/cron"
)

type watcher struct {
	organization string
	repositories []string
	githubToken  string
	biblio       *github.Biblio
	sync.Once
}

type job struct {
	*watcher
	cron *cron.Cron
}
