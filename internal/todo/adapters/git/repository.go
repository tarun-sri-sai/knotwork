package git

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"knotwork/internal/todo/domain"
	"knotwork/internal/todo/ports"

	"github.com/go-git/go-git/v5"
)

const dateFmt = "2006-01-02"
const todoFile = "to-do.txt"

type GitRepository struct {
	gitRepo *git.Repository
	history []historyEntry
}

func getHistory(gitRepo *git.Repository) ([]historyEntry, error) {
	iter, err := gitRepo.Log(&git.LogOptions{
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

func NewGitRepository(repoPath string) (ports.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	history, err := getHistory(repo)
	if err != nil {
		return nil, fmt.Errorf("get repo history: %w", err)
	}

	return &GitRepository{gitRepo: repo, history: history}, nil
}

func (r *GitRepository) getHistoryBetween(startDate, endDate time.Time) ([]historyEntry, error) {
	start, end := -1, -1

	if len(r.history) == 0 {
		return []historyEntry{}, fmt.Errorf("no history")
	}

	if startDate.IsZero() {
		start = 0
	} else {
		start = sort.Search(len(r.history), func(i int) bool {
			return !r.history[i].date.Before(startDate)
		})
		if start == len(r.history) {
			return []historyEntry{}, fmt.Errorf("no commits found from %s", startDate)
		}
	}

	if endDate.IsZero() {
		end = len(r.history) - 1
	} else {
		end := sort.Search(len(r.history), func(i int) bool {
			return r.history[i].date.After(endDate)
		}) - 1
		if end < 0 {
			return []historyEntry{}, fmt.Errorf("no commits found before %s", endDate)
		}
	}

	if start > end {
		return []historyEntry{}, fmt.Errorf("no commits in range")
	}

	return r.history[start : end+1], nil
}

func (r *GitRepository) getTaskMapRecords(startDate, endDate time.Time) ([]taskMapRecord, error) {
	historySlice, err := r.getHistoryBetween(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get task history in date range: %s", err)
	}

	result := make([]taskMapRecord, 0, len(historySlice))

	for _, h := range historySlice {
		commit := h.commit

		file, err := commit.File(todoFile)
		if err != nil {
			return nil, fmt.Errorf("get file from commit %s: %w", commit.Hash, err)
		}

		reader, err := file.Blob.Reader()
		if err != nil {
			return nil, fmt.Errorf("get reader for file in commit %s: %w", commit.Hash, err)
		}

		text, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil, fmt.Errorf("read file content in commit %s: %w", commit.Hash, err)
		}

		blockData, err := ParseTodo(string(text))
		if err != nil {
			return nil, fmt.Errorf("parse todo file in commit %s: %w", commit.Hash, err)
		}

		result = append(result, taskMapRecord{taskMap: blockData, date: h.date})
	}

	return result, nil
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

func (r *GitRepository) GetTaskDurationsBetween(startDateStr, endDateStr string) ([]domain.TaskDuration, error) {
	startDate, err := r.ParseDate(startDateStr)
	if err != nil {
		return []domain.TaskDuration{}, fmt.Errorf("parse start date: %w", err)
	}

	endDate, err := r.ParseDate(endDateStr)
	if err != nil {
		return []domain.TaskDuration{}, fmt.Errorf("parse end date: %w", err)
	}

	taskMapRecords, err := r.getTaskMapRecords(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get task map records: %s", err)
	}

	tasks := make(map[domain.TaskId]domain.TaskDuration)
	for _, taskMapRecord := range taskMapRecords {
		commitDate := taskMapRecord.date

		currTasks := make(map[domain.TaskId]bool)
		for taskID := range taskMapRecord.taskMap {
			currTasks[taskID] = true
		}

		for taskID, task := range taskMapRecord.taskMap {
			if taskDuration, exists := tasks[taskID]; exists {
				taskDuration.Updates = task.Updates
				taskDuration.Category = task.Category
				taskDuration.ParentTasks = task.ParentTasks
				taskDuration.Finished = task.Finished

				if task.Finished && taskDuration.EndDate.IsZero() {
					taskDuration.EndDate = commitDate
				}

				tasks[taskID] = taskDuration
			} else {
				tasks[taskID] = domain.TaskDuration{
					Task: domain.Task{
						Id:          taskID,
						Title:       task.Title,
						Updates:     task.Updates,
						Finished:    task.Finished,
						Category:    task.Category,
						ParentTasks: task.ParentTasks,
					},
					StartDate: commitDate,
				}
			}
		}

		for taskID, taskDuration := range tasks {
			if !currTasks[taskID] && !taskDuration.Finished && taskDuration.EndDate.IsZero() {
				taskDuration.EndDate = commitDate
				tasks[taskID] = taskDuration
			}
		}
	}

	result := make([]domain.TaskDuration, 0, len(tasks))
	for _, td := range tasks {
		result = append(result, td)
	}

	return result, nil
}
