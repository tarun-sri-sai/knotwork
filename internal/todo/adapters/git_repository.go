package adapters

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"knotwork/internal/todo/domain"
	"knotwork/internal/todo/ports"

	"github.com/libgit2/git2go/v34"
)

const dateFmt = "2006-01-02"
const todoFile = "to-do.txt"

type GitRepository struct {
	gitRepo *git.Repository
}

type historyEntry struct {
	date   time.Time
	commit *git.Commit
}

func NewGitRepository(repoPath string) (ports.Repository, error) {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	return &GitRepository{gitRepo: repo}, nil
}

func (r *GitRepository) getHistory() ([]historyEntry, error) {
	walk, err := r.gitRepo.Walk()
	if err != nil {
		return nil, fmt.Errorf("create revwalk: %w", err)
	}
	defer walk.Free()

	if err := walk.PushHead(); err != nil {
		return nil, fmt.Errorf("push HEAD: %w", err)
	}

	walk.Sorting(git.SortTime)

	var result []historyEntry

	err = walk.Iterate(func(commit *git.Commit) bool {
		msg := strings.TrimSpace(commit.Message())

		date, err := time.Parse(dateFmt, msg)
		if err != nil {
			return true
		}

		result = append(result, historyEntry{
			date:   date,
			commit: commit,
		})

		return true
	})
	if err != nil {
		return nil, fmt.Errorf("walk commits: %w", err)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].date.Before(result[j].date)
	})

	return result, nil
}

func (r *GitRepository) getRange(history []historyEntry, startDate, endDate string) (int, int, error) {
	start := -1
	end := -1

	startDateParsed, err := time.Parse(dateFmt, startDate)
	if err != nil {
		start = 0
	}

	endDateParsed, err := time.Parse(dateFmt, endDate)
	if err != nil {
		end = len(history)
	}

	if len(history) == 0 {
		return start, end, fmt.Errorf("no history")
	}

	if start == -1 {
		start = sort.Search(len(history), func(i int) bool {
			return !history[i].date.Before(startDateParsed)
		})
		if start == len(history) {
			return start, end, fmt.Errorf("no commits found from %s", startDate)
		}
	}

	if end == -1 {
		end = sort.Search(len(history), func(i int) bool {
			return history[i].date.After(endDateParsed)
		}) - 1
		if end < 0 {
			return -1, -1, fmt.Errorf("no commits found before %s", endDate)
		}
	}

	if start > end {
		return start, end, fmt.Errorf("no commits in range")
	}

	return start, end, nil
}

func (r *GitRepository) getTaskMaps(history []historyEntry) ([]domain.TaskMap, error) {
	result := make([]domain.TaskMap, 0, len(history))

	for _, h := range history {
		commit := h.commit

		tree, err := commit.Tree()
		if err != nil {
			return nil, fmt.Errorf("get tree for commit %s: %w", commit.Id(), err)
		}
		defer tree.Free()

		entry, err := tree.EntryByPath(todoFile)
		if err != nil {
			return nil, fmt.Errorf("get file from commit %s: %w", commit.Id(), err)
		}

		blob, err := r.gitRepo.LookupBlob(entry.Id)
		if err != nil {
			return nil, fmt.Errorf("lookup blob in commit %s: %w", commit.Id(), err)
		}
		defer blob.Free()

		blockData, err := domain.ParseTodo(string(blob.Contents()))
		if err != nil {
			return nil, fmt.Errorf("parse todo file in commit %s: %w", commit.Id(), err)
		}

		result = append(result, blockData)
	}

	return result, nil
}

func (r *GitRepository) GetTaskDurationsBetween(startDate, endDate string) ([]domain.TaskDuration, error) {
	history, err := r.getHistory()
	if err != nil {
		return nil, err
	}

	startIdx, endIdx, err := r.getRange(history, startDate, endDate)
	if err != nil {
		return nil, err
	}

	historySlice := history[startIdx : endIdx+1]
	taskMaps, err := r.getTaskMaps(historySlice)
	if err != nil {
		return nil, err
	}

	tasks := make(map[domain.TaskId]domain.TaskDuration)
	for i, taskMap := range taskMaps {
		commitDate := historySlice[i].date

		currTasks := make(map[domain.TaskId]bool)
		for taskID := range taskMap {
			currTasks[taskID] = true
		}

		for taskID, task := range taskMap {
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
