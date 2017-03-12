package watcher

import (
	"github.com/aerokite/go-github-watcher/pkg/github"
)

// Compare old and new data
func comapareData(cacheData, newData map[string]*github.RepositoryInfo) {
	/*
		1. Who has hit star button in your repository?
		2. Who has hit unstar button in your repository?
		3. Any new Issue? How many?
		4. Who is watching your repo?
		5. Any new pull request?
		6. Any new release?
		7. Any one fork your repo?
	*/
}
