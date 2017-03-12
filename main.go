package main

import (
	"log"

	"github.com/aerokite/go-github-watcher/pkg/watcher"
)

func main() {
	watcher := watcher.New()

	watcher.SetGithubToken("")

	// Set organization name
	watcher.SetOrganization("")

	// Set repositories. As you like.
	// If you do not add any repository, it will watch all.
	watcher.AddRepositories()

	// Github API call has rate limit. 60 calls per hours.
	// Authorized APP has higher rate limit.
	job, err := watcher.Schedule("@every 30m")
	if err != nil {
		log.Fatalln(err)
	}

	job.RunAndHold()
}
