package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	coreDomain "github.com/tarun-sri-sai/knotwork/internal/core/domain"

	"github.com/spf13/cobra"
)

type TodosInput struct {
	StartDate string `json:"startDate,omitempty" jsonschema:"date from when to find to-do's"`
	EndDate   string `json:"endDate,omitempty"   jsonschema:"date till when to find to-do's"`
	MinDays   int    `json:"minDays,omitempty"   jsonschema:"minimum age of a finished/abandoned to-do for it to be included in the result"`
	Type      string `json:"type,omitempty"      jsonschema:"type of task (\"\", \"abandoned\", \"finished\")"`
}

type TodosOutput struct {
	TaskInfo coreDomain.TaskInfo `json:"taskInfo" jsonschema:"task info containing the stats and task details"`
}

var core *Core

func Todos(todosInput TodosInput) (TodosOutput, error) {
	endDate, err := core.repository.ParseDate(todosInput.EndDate)
	if err != nil {
		return TodosOutput{}, fmt.Errorf("parse end date: %w", err)
	}

	tasks, err := core.repository.GetTasksBetween(todosInput.StartDate, todosInput.EndDate)
	if err != nil {
		return TodosOutput{}, fmt.Errorf("get task data: %w", err)
	}

	var taskInfo coreDomain.TaskInfo

	taskType := coreDomain.TaskType(todosInput.Type)
	switch taskType {
	case coreDomain.TaskTypeFinished:
		taskInfo, err = coreDomain.GetFinishedTaskInfoBetween(tasks, endDate, todosInput.MinDays)
		if err != nil {
			return TodosOutput{}, fmt.Errorf("get finished task info: %w", err)
		}
	case coreDomain.TaskTypeAbandoned:
		taskInfo, err = coreDomain.GetAbandonedTaskInfoBetween(tasks, endDate, todosInput.MinDays)
		if err != nil {
			return TodosOutput{}, fmt.Errorf("get abandoned task info: %w", err)
		}
	case coreDomain.TaskTypeAll:
		taskInfo, err = coreDomain.GetTaskInfoBetween(tasks, endDate, todosInput.MinDays)
		if err != nil {
			return TodosOutput{}, fmt.Errorf("get task info: %w", err)
		}
	default:
		return TodosOutput{}, errors.New("invalid task type")
	}

	return TodosOutput{TaskInfo: taskInfo}, nil
}

var (
	repoType  string
	repoDsn   string
	taskType  string
	startDate string
	endDate   string
	minDays   int
)

var rootCmd = &cobra.Command{
	Use:   "knotwork-cli",
	Short: "Knotwork CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		core, err = NewCore(repoType, repoDsn)
		if err != nil {
			return err
		}

		todosInput := TodosInput{
			StartDate: startDate,
			EndDate:   endDate,
			MinDays:   minDays,
			Type:      taskType,
		}

		todosOutput, err := Todos(todosInput)

		data, err := json.MarshalIndent(todosOutput, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal todos output: %w", err)
		}

		fmt.Println(string(data))

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&repoType, "repo-type", "t", "", "Repo Type")
	rootCmd.Flags().StringVarP(&repoDsn, "repo-dsn", "d", "", "Repo DSN")
	rootCmd.Flags().StringVarP(&taskType, "task-type", "y", "", "Task Type (\"\", \"abandoned\", \"finished\")")
	rootCmd.Flags().StringVarP(&startDate, "start-date", "s", "", "Start Date (YYYY-MM-DD)")
	rootCmd.Flags().StringVarP(&endDate, "end-date", "e", "", "End Date (YYYY-MM-DD)")
	rootCmd.Flags().IntVarP(&minDays, "min-days", "m", 0, "Minimum Days")

	rootCmd.MarkFlagRequired("repo-type")
	rootCmd.MarkFlagRequired("repo-dsn")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
