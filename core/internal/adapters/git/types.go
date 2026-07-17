package git

import (
	"time"

	"knotwork-core/internal/domain"
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
