package adapters

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"knotwork-core/internal/todo/domain"
	"knotwork-core/internal/todo/ports"
)

const dateFmt = "2006-01-02"
const todoFile = "to-do.txt"

type GitRepository struct {
	gitRepo *git.Repository
}

type historyEntry struct {
	date   time.Time
	commit *object.Commit
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
			fmt.Printf("commit %s - invalid date format: %s", c.Hash, msg)
		}

		result = append(result, historyEntry{date: date, commit: c})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].date.Before(result[j].date)
	})
	return result, nil
}

func (r *GitRepository) getRange(startDate, endDate string) (int, int, error) {
	startDateParsed, err := time.Parse(dateFmt, startDate)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid start date format: %w", err)
	}

	endDateParsed, err := time.Parse(dateFmt, endDate)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid end date format: %w", err)
	}

	history, err := r.getHistory()
	if err != nil {
		return -1, -1, err
	}

	if len(history) == 0 {
		return 0, 0, fmt.Errorf("no history")
	}

	start := sort.Search(len(history), func(i int) bool {
		return history[i].date.Before(startDateParsed)
	})
	if start == len(history) {
		return 0, 0, fmt.Errorf("no commits found from %s", startDate)
	}

	end := sort.Search(len(history), func(i int) bool {
		return history[i].date.After(endDateParsed)
	}) - 1
	if end < 0 {
		return 0, 0, fmt.Errorf("no commits found before %s", endDate)
	}

	if start > end {
		return 0, 0, fmt.Errorf("no commits in range")
	}
	return start, end, nil
}

func (r *GitRepository) getTaskMaps(startIdx, endIdx int) ([]domain.TaskMap, error) {
	history, err := r.getHistory()
	if err != nil {
		return nil, err
	}

	result := []domain.TaskMap{}
	for i := startIdx; i <= endIdx; i++ {
		c := history[i].commit
		file, err := c.File(todoFile)
		if err != nil {
			return nil, fmt.Errorf("get file from commit %s: %w", c.Hash, err)
		}

		reader, err := file.Blob.Reader()
		if err != nil {
			return nil, fmt.Errorf("get reader for file in commit %s: %w", c.Hash, err)
		}
		defer reader.Close()

		text, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("read file content in commit %s: %w", c.Hash, err)
		}

		blockData, err := domain.ParseTodo(string(text))
		if err != nil {
			return nil, fmt.Errorf("parse todo file in commit %s: %w", c.Hash, err)
		}

		result = append(result, blockData)
	}

	return result, nil
}

func (r *GitRepository) GetTaskMapBefore(date string) ([]domain.TaskMap, error) {
	history, err := r.getHistory()
	if err != nil {
		return nil, err
	}

	startIdx, endIdx, err := r.getRange(history[0].date.Format(dateFmt), date)
	if err != nil {
		return nil, err
	}

	result, err := r.getTaskMaps(startIdx, endIdx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *GitRepository) GetTaskMapBetween(startDate, endDate string) ([]domain.TaskMap, error) {
	startIdx, endIdx, err := r.getRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	result, err := r.getTaskMaps(startIdx, endIdx)
	if err != nil {
		return nil, err
	}

	return result, nil
}
