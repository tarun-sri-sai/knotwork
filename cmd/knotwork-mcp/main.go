package main

import (
	"context"
	"fmt"
	"log"

	coreDomain "github.com/tarun-sri-sai/knotwork/internal/core/domain"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

type TodosInput struct {
	StartDate string `json:"startDate,omitempty" jsonschema:"date from when to find to-do's"`
	EndDate   string `json:"endDate,omitempty" jsonschema:"date till when to find to-do's"`
	MinDays   int    `json:"minDays,omitempty" jsonschema:"minimum age of a finished/abandoned to-do for it to be included in the result"`
	Type      string `json:"type,omitempty" jsonschema:"type of task (\"\", \"abandoned\", \"finished\")"`
}

type TodosOutput struct {
	TaskInfo coreDomain.TaskInfo `json:"taskInfo" jsonschema:"task info containing the stats and task details"`
}

var core *Core

func Todos(ctx context.Context, req *mcp.CallToolRequest, todosInput TodosInput) (*mcp.CallToolResult, TodosOutput, error) {
	endDate, err := core.repository.ParseDate(todosInput.EndDate)
	if err != nil {
		return nil, TodosOutput{}, fmt.Errorf("parse end date: %w", err)
	}

	tasks, err := core.repository.GetTasksBetween(todosInput.StartDate, todosInput.EndDate)
	if err != nil {
		return nil, TodosOutput{}, fmt.Errorf("get task data: %w", err)
	}

	var taskInfo coreDomain.TaskInfo

	taskType := coreDomain.TaskType(todosInput.Type)
	switch taskType {
	case coreDomain.TaskTypeFinished:
		taskInfo, err = coreDomain.GetFinishedTaskInfoBetween(tasks, endDate, todosInput.MinDays)
		if err != nil {
			return nil, TodosOutput{}, fmt.Errorf("get finished task info: %w", err)
		}
	case coreDomain.TaskTypeAbandoned:
		taskInfo, err = coreDomain.GetAbandonedTaskInfoBetween(tasks, endDate, todosInput.MinDays)
		if err != nil {
			return nil, TodosOutput{}, fmt.Errorf("get abandoned task info: %w", err)
		}
	case coreDomain.TaskTypeAll:
		taskInfo, err = coreDomain.GetTaskInfoBetween(tasks, endDate, todosInput.MinDays)
		if err != nil {
			return nil, TodosOutput{}, fmt.Errorf("get task info: %w", err)
		}
	default:
		return nil, TodosOutput{}, fmt.Errorf("invalid task type")
	}

	return &mcp.CallToolResult{}, TodosOutput{TaskInfo: taskInfo}, nil
}

var (
	repoType string
	repoDsn  string
)

var rootCmd = &cobra.Command{
	Use:   "knotwork-mcp",
	Short: "Knotwork MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		core, err = NewCore(repoType, repoDsn)
		if err != nil {
			return err
		}

		server := mcp.NewServer(
			&mcp.Implementation{
				Name:    "knotwork",
				Version: "v1.0.0",
			},
			nil,
		)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "todos",
			Description: "get todos of the user",
		}, Todos)

		return server.Run(context.Background(), &mcp.StdioTransport{})
	},
}

func init() {
	rootCmd.Flags().StringVarP(&repoType, "repo-type", "t", "", "Repo Type")
	rootCmd.Flags().StringVarP(&repoDsn, "repo-dsn", "d", "", "Repo DSN")

	rootCmd.MarkFlagRequired("repo-type")
	rootCmd.MarkFlagRequired("repo-dsn")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
