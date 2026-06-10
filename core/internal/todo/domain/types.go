package domain

type category struct {
	Id   string	`json:"id"`
	Category string	`json:"category"`
}

type task struct {
	Id	string	`json:"id"`
	Level int	`json:"level"`
	Title string	`json:"title"`
	Updates []string `json:"updates"`
	Finished bool `json:"finished"`
}

type block interface {
    isBlock()
}

func (category) isBlock() {}
func (task) isBlock() {}

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
