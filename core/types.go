package main

type TaskId string

type Task struct {
	Id          TaskId   `json:"id"`
	Title       string   `json:"title"`
	Updates     []string `json:"updates"`
	Finished    bool     `json:"finished"`
	Category    *string   `json:"category"`
	ParentTasks []string `json:"parentTasks"`
}

type TaskMap map[TaskId]Task
