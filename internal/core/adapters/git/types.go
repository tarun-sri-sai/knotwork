package git

import (
	"time"

	"github.com/tarun-sri-sai/knotwork/internal/core/domain"
)

type parsedTask struct {
	id          domain.TaskId
	title       string
	updates     []string
	finished    bool
	category    string
	parentTasks []string
}

type parsedTaskMap map[domain.TaskId]parsedTask

type parsedTaskMapDated struct {
	taskMap parsedTaskMap
	date    time.Time
}
