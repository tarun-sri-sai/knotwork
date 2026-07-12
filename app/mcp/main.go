package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"knotwork/internal/todo/domain"
)

const coreBaseURL = "http://core:80"

type TodosInput struct {
	StartDate string          `json:"startDate,omitempty" jsonschema:"date from when to find to-do's"`
	EndDate   string          `json:"endDate,omitempty" jsonschema:"date till when to find to-do's"`
	MinDays   int             `json:"minDays,omitempty" jsonschema:"minimum age of a finished/abandoned to-do for it to be included in the result"`
	Type      domain.TaskType `json:"type,omitempty" jsonschema:"type of task (\"\", \"abandoned\", \"finished\")"`
}

type TodosOutput struct {
	TaskInfo domain.TaskInfo `json:"taskInfo" jsonschema:"task info containing the stats and task details"`
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

	var taskInfo domain.TaskInfo
	if err := json.NewDecoder(resp.Body).Decode(&taskInfo); err != nil {
		return nil, TodosOutput{}, fmt.Errorf("decode core todos response: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "retrieved todos"}},
	}, TodosOutput{TaskInfo: taskInfo}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "knotwork", Version: "v0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "todos", Description: "get todos of the user"}, Todos)

	transport := &mcp.StreamableServerTransport{}
	go func() {
		if err := server.Run(context.Background(), transport); err != nil {
			log.Fatal(err)
		}
	}()
	
	http.Handle("/mcp", transport)

	port := 80
	if s, ok := os.LookupEnv("PORT"); ok && s != "" {
		if p, err := strconv.Atoi(s); err == nil && p > 0 && p <= 0xffff {
			port = p
		}
	}

	log.Printf("serving on port %d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), transport); err != nil {
		log.Fatal(err)
	}
}
