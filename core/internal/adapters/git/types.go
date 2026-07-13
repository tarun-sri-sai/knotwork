package git

import (
	"time"

	"knotwork-core/internal/domain"

	"github.com/go-git/go-git/v5/plumbing/object"
)

type historyEntry struct {
	date   time.Time
	commit *object.Commit
}

type taskMap map[domain.TaskId]domain.Task

type taskMapRecord struct {
	taskMap taskMap
	date    time.Time
}

type category struct {
	id       string
	category string
}

type task struct {
	id       string
	level    int
	title    string
	updates  []string
	finished bool
}

type block interface {
	isBlock()
}

func (category) isBlock() {}
func (task) isBlock()     {}
