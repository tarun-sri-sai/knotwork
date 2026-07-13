package domain

import (
	"fmt"
	"slices"
	"time"
)

func getFinishedTaskDurations(taskDurations []TaskDuration) []TaskDuration {
	var finishedTasks []TaskDuration
	for _, task := range taskDurations {
		if task.Finished {
			finishedTasks = append(finishedTasks, task)
		}
	}

	return finishedTasks
}

func getAbandonedTaskDurations(taskDurations []TaskDuration) []TaskDuration {
	var abandonedTasks []TaskDuration
	for _, task := range taskDurations {
		if !task.Finished && !task.EndDate.IsZero() {
			abandonedTasks = append(abandonedTasks, task)
		}
	}

	return abandonedTasks
}

func getTaskDurationsByMinDays(taskDurations []TaskDuration, minDays int) ([]TaskDuration, error) {
	if minDays < 0 {
		return nil, fmt.Errorf("minimum days cannot be negative")
	}

	if minDays == 0 {
		return taskDurations, nil
	}

	var result []TaskDuration
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

func getTaskStats(taskDurations []TaskDuration, endDate time.Time) TaskStats {
	if len(taskDurations) == 0 {
		return TaskStats{
			TotalTasks:          0,
			LongestTaskId:       "",
			AverageTaskDuration: 0,
			MedianTaskDuration:  0,
			MostActiveTaskId:    "",
			MostActiveCategory:  "",
		}
	}

	type taskData struct {
		taskId   TaskId
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

	return TaskStats{
		TotalTasks:          len(tasks),
		LongestTaskId:       longestTask.taskId,
		AverageTaskDuration: totalDuration / len(tasks),
		MedianTaskDuration:  medianTask.duration,
		MostActiveTaskId:    mostActiveTask.taskId,
		MostActiveCategory:  mostActiveCategory,
	}
}

func GetTaskInfoBetween(taskDurations []TaskDuration, endDate time.Time, minDays int) (TaskInfo, error) {
	filteredTaskDurations, err := getTaskDurationsByMinDays(taskDurations, minDays)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("filter task durations by min days: %w", err)
	}

	taskStats := getTaskStats(filteredTaskDurations, endDate)

	return TaskInfo{
		TaskStats:     taskStats,
		TaskDurations: filteredTaskDurations,
	}, nil
}

func GetFinishedTaskInfoBetween(taskDurations []TaskDuration, endDate time.Time, minDays int) (TaskInfo, error) {
	filteredTaskDurations := getFinishedTaskDurations(taskDurations)

	filteredTaskDurations, err := getTaskDurationsByMinDays(filteredTaskDurations, minDays)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("filter task durations by min days: %w", err)
	}

	taskStats := getTaskStats(filteredTaskDurations, endDate)

	return TaskInfo{
		TaskStats:     taskStats,
		TaskDurations: filteredTaskDurations,
	}, nil
}

func GetAbandonedTaskInfoBetween(taskDurations []TaskDuration, endDate time.Time, minDays int) (TaskInfo, error) {
	filteredTaskDurations := getAbandonedTaskDurations(taskDurations)

	filteredTaskDurations, err := getTaskDurationsByMinDays(filteredTaskDurations, minDays)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("filter task durations by min days: %w", err)
	}

	taskStats := getTaskStats(filteredTaskDurations, endDate)

	return TaskInfo{
		TaskStats:     taskStats,
		TaskDurations: filteredTaskDurations,
	}, nil
}
