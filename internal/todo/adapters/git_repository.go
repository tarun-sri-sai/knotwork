package adapters

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

	"knotwork/internal/todo/domain"
	"knotwork/internal/todo/ports"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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
			continue
		}

		result = append(result, historyEntry{date: date, commit: c})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].date.Before(result[j].date)
	})
	return result, nil
}

func (r *GitRepository) getRange(history []historyEntry, startDate, endDate time.Time) (int, int, error) {
	if len(history) == 0 {
		return -1, -1, fmt.Errorf("no history")
	}

	start := sort.Search(len(history), func(i int) bool {
		return !history[i].date.Before(startDate)
	})
	if start == len(history) {
		return -1, -1, fmt.Errorf("no commits found from %s", startDate)
	}

	end := sort.Search(len(history), func(i int) bool {
		return history[i].date.After(endDate)
	}) - 1
	if end < 0 {
		return -1, -1, fmt.Errorf("no commits found before %s", endDate)
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

		blockData, err := domain.ParseTodo(string(text))
		if err != nil {
			return nil, fmt.Errorf("parse todo file in commit %s: %w", commit.Hash, err)
		}

		result = append(result, blockData)
	}

	return result, nil
}

func (r *GitRepository) getTaskDurationsBetween(startDate, endDate time.Time) ([]domain.TaskDuration, error) {
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

func (r *GitRepository) getFinishedTaskDurations(taskDurations []domain.TaskDuration) ([]domain.TaskDuration) {
	var finishedTasks []domain.TaskDuration
	for _, task := range taskDurations {
		if task.Finished {
			finishedTasks = append(finishedTasks, task)
		}
	}

	return finishedTasks
}

func (r *GitRepository) getAbandonedTaskDurations(taskDurations []domain.TaskDuration) ([]domain.TaskDuration) {
	var abandonedTasks []domain.TaskDuration
	for _, task := range taskDurations {
		if !task.Finished && !task.EndDate.IsZero() {
			abandonedTasks = append(abandonedTasks, task)
		}
	}

	return abandonedTasks
}

func (r *GitRepository) getTaskDurationsByMinDays(taskDurations []domain.TaskDuration, minDays int) ([]domain.TaskDuration, error) {
	if (minDays < 0) {
		return nil, fmt.Errorf("minimum days cannot be negative")
	}

	if (minDays == 0) {
		return taskDurations, nil
	}

	var result []domain.TaskDuration
	for _, task := range taskDurations {
		if task.EndDate.IsZero() {
			continue
		}

		duration := task.EndDate.Sub(task.StartDate)
		if int64(duration.Hours()) >= 24*int64(minDays) {
			result = append(result, task)
		}
	}

	return result, nil
}

func (r *GitRepository) getTaskStats(taskDurations []domain.TaskDuration, endDate time.Time) domain.TaskStats {
	if len(taskDurations) == 0 {
		return domain.TaskStats{
			TotalTasks:          0,
			LongestTaskId:       "",
			AverageTaskDuration: 0,
			MedianTaskDuration:  0,
			MostActiveTaskId:    "",
			MostActiveCategory:  "",
		}
	}

	type taskData struct {
		taskId   domain.TaskId
		duration int
		updates  int
		category string
	}

	tasks := make([]taskData, 0, len(taskDurations))

	for _, task := range taskDurations {
		taskEndDate := task.EndDate
		if taskEndDate.IsZero() {
			taskEndDate = endDate
		}

		duration := int(taskEndDate.Sub(task.StartDate).Hours() / 24)

		tasks = append(tasks, taskData{
			taskId:   task.Id,
			duration: duration,
			updates:  len(task.Updates),
			category: task.Category,
		})
	}

	categoryUpdates := make(map[string]int)

	for _, task := range tasks {
		categoryUpdates[task.category] += task.updates
	}

	longestTask := tasks[0]
	mostActiveTask := tasks[0]
	totalDuration := 0

	for _, task := range tasks {
		if task.duration > longestTask.duration {
			longestTask = task
		}

		if task.updates > mostActiveTask.updates {
			mostActiveTask = task
		}

		totalDuration += task.duration
	}

	slices.SortFunc(tasks, func(a, b taskData) int {
		return a.duration - b.duration
	})

	medianTask := tasks[len(tasks)/2]

	mostActiveCategory := ""
	maxUpdates := 0

	for category, updates := range categoryUpdates {
		if updates > maxUpdates {
			maxUpdates = updates
			mostActiveCategory = category
		}
	}

	return domain.TaskStats{
		TotalTasks:          len(tasks),
		LongestTaskId:       longestTask.taskId,
		AverageTaskDuration: totalDuration / len(tasks),
		MedianTaskDuration:  medianTask.duration,
		MostActiveTaskId:    mostActiveTask.taskId,
		MostActiveCategory:  mostActiveCategory,
	}
}

func (r *GitRepository) GetTaskInfoBetween(startDate, endDate string, minDays int) (domain.TaskInfo, error) {
	start, err := time.Parse(dateFmt, startDate)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("parse start date: %w", err)
	}

	end, err := time.Parse(dateFmt, endDate)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("parse end date: %w", err)
	}

	taskDurations, err := r.getTaskDurationsBetween(start, end)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("get task durations: %w", err)
	}

	filteredTaskDurations, err := r.getTaskDurationsByMinDays(taskDurations, minDays)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("filter task durations by min days: %w", err)
	}

	taskStats := r.getTaskStats(filteredTaskDurations, end)

	return domain.TaskInfo{
		TaskStats:     taskStats,
		TaskDurations: filteredTaskDurations,
	}, nil
}

func (r *GitRepository) GetFinishedTaskInfoBetween(startDate, endDate string, minDays int) (domain.TaskInfo, error) {
	start, err := time.Parse(dateFmt, startDate)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("parse start date: %w", err)
	}

	end, err := time.Parse(dateFmt, endDate)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("parse end date: %w", err)
	}

	taskDurations, err := r.getTaskDurationsBetween(start, end)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("get task durations: %w", err)
	}

	filteredTaskDurations := r.getFinishedTaskDurations(taskDurations)

	filteredTaskDurations, err = r.getTaskDurationsByMinDays(filteredTaskDurations, minDays)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("filter task durations by min days: %w", err)
	}

	taskStats := r.getTaskStats(filteredTaskDurations, end)

	return domain.TaskInfo{
		TaskStats:     taskStats,
		TaskDurations: filteredTaskDurations,
	}, nil
}

func (r *GitRepository) GetAbandonedTaskInfoBetween(startDate, endDate string, minDays int) (domain.TaskInfo, error) {
	start, err := time.Parse(dateFmt, startDate)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("parse start date: %w", err)
	}

	end, err := time.Parse(dateFmt, endDate)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("parse end date: %w", err)
	}

	taskDurations, err := r.getTaskDurationsBetween(start, end)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("get task durations: %w", err)
	}

	filteredTaskDurations := r.getAbandonedTaskDurations(taskDurations)

	filteredTaskDurations, err = r.getTaskDurationsByMinDays(filteredTaskDurations, minDays)
	if err != nil {
		return domain.TaskInfo{}, fmt.Errorf("filter task durations by min days: %w", err)
	}

	taskStats := r.getTaskStats(filteredTaskDurations, end)

	return domain.TaskInfo{
		TaskStats:     taskStats,
		TaskDurations: filteredTaskDurations,
	}, nil
}
