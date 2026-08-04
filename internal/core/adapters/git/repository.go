package git

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/tarun-sri-sai/knotwork/internal/core/domain"
	"github.com/tarun-sri-sai/knotwork/internal/core/ports"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const dateFmt = "2006-01-02"
const todoFile = "to-do.txt"

type historyEntry struct {
	date   time.Time
	commit *object.Commit
}

type GitRepository struct {
	gitRepo *git.Repository
}

func NewGitRepository(repoPath string) (ports.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	return &GitRepository{gitRepo: repo}, nil
}

func (r *GitRepository) getHistory() ([]historyEntry, error) {
	iter, err := r.gitRepo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	result := []historyEntry{}
	for {
		c, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, nil
		}

		msg := strings.TrimSpace(c.Message)
		date, err := time.Parse(dateFmt, msg)
		if err != nil {
			continue
		}

		result = append(result, historyEntry{date: date, commit: c})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].date.Before(result[j].date)
	})
	return result, nil
}

func (r *GitRepository) getHistoryBetween(startDate, endDate time.Time) ([]historyEntry, error) {
	var start, end int

	history, err := r.getHistory()
	if err != nil {
		return nil, fmt.Errorf("get repo history: %w", err)
	}

	if len(history) == 0 {
		return []historyEntry{}, fmt.Errorf("no history")
	}

	if startDate.IsZero() {
		start = 0
	} else {
		start = sort.Search(len(history), func(i int) bool {
			return !history[i].date.Before(startDate)
		})
		if start == len(history) {
			return []historyEntry{}, fmt.Errorf("no commits found from %s", startDate)
		}
	}

	if endDate.IsZero() {
		end = len(history) - 1
	} else {
		end = sort.Search(len(history), func(i int) bool {
			return history[i].date.After(endDate)
		}) - 1
		if end < 0 {
			return []historyEntry{}, fmt.Errorf("no commits found before %s", endDate)
		}
	}

	if start > end {
		return []historyEntry{}, fmt.Errorf("no commits in range")
	}

	return history[start : end+1], nil
}

func (r *GitRepository) getTaskMapDated(historyEntry historyEntry) (parsedTaskMapDated, error) {
	commit := historyEntry.commit

	file, err := commit.File(todoFile)
	if err != nil {
		return parsedTaskMapDated{}, fmt.Errorf("get file from commit %s: %w", commit.Hash, err)
	}

	reader, err := file.Reader()
	if err != nil {
		return parsedTaskMapDated{}, fmt.Errorf("get reader for file in commit %s: %w", commit.Hash, err)
	}

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("failed to close reader: %s\n", err.Error())
		}
	}()

	text, err := io.ReadAll(reader)
	if err != nil {
		return parsedTaskMapDated{}, fmt.Errorf("read file content in commit %s: %w", commit.Hash, err)
	}

	blockData, err := ParseTodo(string(text))
	if err != nil {
		return parsedTaskMapDated{}, fmt.Errorf("parse todo file in commit %s: %w", commit.Hash, err)
	}

	return parsedTaskMapDated{taskMap: blockData, date: historyEntry.date}, nil
}

func (r *GitRepository) ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(dateFmt, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse start date: %w", err)
	}

	return parsed, nil
}

func (r *GitRepository) GetTasksBetween(startDateStr, endDateStr string) ([]domain.Task, error) {
	startDate, err := r.ParseDate(startDateStr)
	if err != nil {
		return []domain.Task{}, fmt.Errorf("parse start date: %w", err)
	}

	endDate, err := r.ParseDate(endDateStr)
	if err != nil {
		return []domain.Task{}, fmt.Errorf("parse end date: %w", err)
	}

	historySlice, err := r.getHistoryBetween(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get task history in date range: %w", err)
	}

	tasks := make(map[domain.TaskId]domain.Task)
	for _, h := range historySlice {
		taskMapDated, err := r.getTaskMapDated(h)
		if err != nil {
			log.Printf("get task map dated for %s: %s\n", h.commit, err.Error())
			continue
		}

		commitDate := taskMapDated.date

		currTasks := make(map[domain.TaskId]bool)
		for taskID := range taskMapDated.taskMap {
			currTasks[taskID] = true
		}

		for taskID, parsedTask := range taskMapDated.taskMap {
			var task domain.Task

			if existingTask, exists := tasks[taskID]; exists && existingTask.EndDate.IsZero() {
				task = existingTask
				task.Updates = parsedTask.updates
				task.Category = parsedTask.category
				task.ParentTasks = parsedTask.parentTasks
				task.Finished = parsedTask.finished
			} else {
				task = domain.Task{
					Id:          taskID,
					Title:       parsedTask.title,
					Updates:     parsedTask.updates,
					Finished:    parsedTask.finished,
					Category:    parsedTask.category,
					ParentTasks: parsedTask.parentTasks,
					StartDate:   commitDate,
				}
			}

			if parsedTask.finished {
				task.EndDate = commitDate
			}

			tasks[taskID] = task
		}

		for taskID, task := range tasks {
			if !currTasks[taskID] && !task.Finished && task.EndDate.IsZero() {
				task.EndDate = commitDate
				tasks[taskID] = task
			}
		}
	}

	result := make([]domain.Task, 0, len(tasks))
	for _, td := range tasks {
		result = append(result, td)
	}

	return result, nil
}
