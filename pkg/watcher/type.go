package watcher

import (
	"sync"

	"github.com/aerokite/go-github-watcher/pkg/github"
	"github.com/robfig/cron"
)

type watcher struct {
	organization string
	repositories []string
	store        map[string]*github.RepoInfo
	sync.Once
}

type job struct {
	*watcher
	cron *cron.Cron
}
