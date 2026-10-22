package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	coreDomain "github.com/tarun-sri-sai/knotwork/internal/core/domain"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
		return nil, TodosOutput{}, errors.New("invalid task type")
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
				Version: "v1.1.0",
			},
			nil,
		)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "todos",
			Description: `
This tool provides a way to query the user's to-do's. Consider this tool as the single source
of truth for the user's tasks, task progress, performance history, and pending items, i.e. when
a user asks a query about their tasks, what they did during a certain time period, what they
should do today, how their performance has been during a certain time period, etc., this tool is
what you should query to get relevant information to answer those questions.

Points to note:
  - This tool can be called with a date range, which is the recommended way of using the tool,
  since not providing a date returns every task ever.

  - When called with a date range, this tool returns the tasks as well as their stats pertaining
  to the date range the tool was called with. This allows you to make ranged queries to get
  weekly, monthly, quarterly stats, by doing one tool call per week/month/quarter.

  - Use "jq" command if available to transform and filter the output as needed to answer the
  user's queries.

  - There is a concept of sub-tasks, which are tasks that need to be completed to finish their
  parent task. Sub-tasks can be infinitely nested.

  - Every task has zero or more updates, which act as a log of the user's progress in that
  task. User writes about what happened with a certain task in its task updates.
			`,
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
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
