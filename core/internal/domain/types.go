package domain

import (
	"time"
)

type TaskId string

type Task struct {
	Id          TaskId    `json:"id"`
	Title       string    `json:"title"`
	Updates     []string  `json:"updates"`
	Finished    bool      `json:"finished"`
	Category    string    `json:"category"`
	ParentTasks []string  `json:"parentTasks"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate,omitzero"`
}

type Stats struct {
	TotalTasks          int    `json:"totalTasks"`
	LongestTaskId       TaskId `json:"longestTaskId"`
	AverageTaskDuration int    `json:"averageTaskDuration"`
	MedianTaskDuration  int    `json:"medianTaskDuration"`
	MostActiveTaskId    TaskId `json:"mostActiveTaskId"`
	MostActiveCategory  string `json:"mostActiveCategory"`
}

type TaskInfo struct {
	Stats Stats  `json:"stats"`
	Tasks []Task `json:"tasks"`
}

type TaskType string

const (
	TaskTypeFinished  TaskType = "finished"
	TaskTypeAbandoned TaskType = "abandoned"
	TaskTypeAll       TaskType = ""
)
