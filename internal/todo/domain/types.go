package domain

import (
	"time"
)

type category struct {
	Id       string `json:"id"`
	Category string `json:"category"`
}

type task struct {
	Id       string   `json:"id"`
	Level    int      `json:"level"`
	Title    string   `json:"title"`
	Updates  []string `json:"updates"`
	Finished bool     `json:"finished"`
}

type block interface {
	isBlock()
}

func (category) isBlock() {}
func (task) isBlock()     {}

type TaskId string

type Task struct {
	Id          TaskId   `json:"id"`
	Title       string   `json:"title"`
	Updates     []string `json:"updates"`
	Finished    bool     `json:"finished"`
	Category    string   `json:"category"`
	ParentTasks []string `json:"parentTasks"`
}

type TaskMap map[TaskId]Task

type TaskDuration struct {
	Task
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type TaskStats struct {
	TotalTasks          int    `json:"totalTasks"`
	LongestTaskId       TaskId `json:"longestTaskId"`
	AverageTaskDuration int    `json:"averageTaskDuration"`
	MedianTaskDuration  int    `json:"medianTaskDuration"`
	MostActiveTaskId    TaskId `json:"mostActiveTaskId"`
	MostActiveCategory  string `json:"mostActiveCategory"`
}

type TaskInfo struct {
	TaskStats     TaskStats      `json:"stats"`
	TaskDurations []TaskDuration `json:"durations"`
}
