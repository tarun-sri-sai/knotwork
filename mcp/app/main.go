package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const coreBaseURL = "http://core:80"

type TodosInput struct {
	StartDate string `json:"startDate,omitempty" jsonschema:"date from when to find to-do's"`
	EndDate   string `json:"endDate,omitempty" jsonschema:"date till when to find to-do's"`
	MinDays   int    `json:"minDays,omitempty" jsonschema:"minimum age of a finished/abandoned to-do for it to be included in the result"`
	Type      string `json:"type,omitempty" jsonschema:"type of task (\"\", \"abandoned\", \"finished\")"`
}

type TodosOutput struct {
	TaskInfo map[string]any `json:"taskInfo" jsonschema:"task info containing the stats and task details"`
}

func Todos(ctx context.Context, req *mcp.CallToolRequest, todosInput TodosInput) (*mcp.CallToolResult, TodosOutput, error) {
	endpoint, err := url.Parse(coreBaseURL + "/todos")
	if err != nil {
		return nil, TodosOutput{}, fmt.Errorf("build core todos URL: %w", err)
	}

	query := endpoint.Query()
	if todosInput.StartDate != "" {
		query.Set("startDate", todosInput.StartDate)
	}
	if todosInput.EndDate != "" {
		query.Set("endDate", todosInput.EndDate)
	}
	if todosInput.MinDays > 0 {
		query.Set("minDays", strconv.Itoa(todosInput.MinDays))
	}
	if todosInput.Type != "" {
		query.Set("type", string(todosInput.Type))
	}
	endpoint.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, TodosOutput{}, fmt.Errorf("create core todos request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, TodosOutput{}, fmt.Errorf("call core todos endpoint: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body reader: %s\n", err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, TodosOutput{}, fmt.Errorf("core todos request failed: %s: %s", resp.Status, string(body))
	}

	var taskInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&taskInfo); err != nil {
		return nil, TodosOutput{}, fmt.Errorf("decode core todos response: %w", err)
	}

	return &mcp.CallToolResult{}, TodosOutput{TaskInfo: taskInfo}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "knotwork", Version: "v0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "todos", Description: "get todos of the user"}, Todos)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
