package domain

import (
	"fmt"
	"slices"
	"time"
)

type taskData struct {
	taskId   TaskId
	duration int
	updates  int
	category string
}

func getFinishedTasks(tasks []Task) []Task {
	var finishedTasks []Task
	for _, task := range tasks {
		if task.Finished {
			finishedTasks = append(finishedTasks, task)
		}
	}

	return finishedTasks
}

func getAbandonedTasks(tasks []Task) []Task {
	var abandonedTasks []Task
	for _, task := range tasks {
		if !task.Finished && !task.EndDate.IsZero() {
			abandonedTasks = append(abandonedTasks, task)
		}
	}

	return abandonedTasks
}

func getTasksByMinDays(tasks []Task, minDays int) ([]Task, error) {
	if minDays < 0 {
		return nil, fmt.Errorf("minimum days cannot be negative")
	}

	if minDays == 0 {
		return tasks, nil
	}

	var result []Task
	for _, task := range tasks {
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

func getTaskStats(tasks []Task, endDate time.Time) Stats {
	if len(tasks) == 0 {
		return Stats{
			TotalTasks:          0,
			LongestTaskId:       "",
			AverageTaskDuration: 0,
			MedianTaskDuration:  0,
			MostActiveTaskId:    "",
			MostActiveCategory:  "",
		}
	}

	taskDataList := make([]taskData, 0, len(tasks))

	for _, task := range tasks {
		taskEndDate := task.EndDate
		if taskEndDate.IsZero() {
			taskEndDate = endDate
		}

		duration := int(taskEndDate.Sub(task.StartDate).Hours() / 24)

		taskDataList = append(taskDataList, taskData{
			taskId:   task.Id,
			duration: duration,
			updates:  len(task.Updates),
			category: task.Category,
		})
	}

	categoryUpdates := make(map[string]int)

	for _, td := range taskDataList {
		categoryUpdates[td.category] += td.updates
	}

	longestTask := taskDataList[0]
	mostActiveTask := taskDataList[0]
	totalDuration := 0

	for _, td := range taskDataList {
		if td.duration > longestTask.duration {
			longestTask = td
		}

		if td.updates > mostActiveTask.updates {
			mostActiveTask = td
		}

		totalDuration += td.duration
	}

	slices.SortFunc(taskDataList, func(a, b taskData) int {
		return a.duration - b.duration
	})

	medianTask := taskDataList[len(taskDataList)/2]

	mostActiveCategory := ""
	maxUpdates := 0

	for category, updates := range categoryUpdates {
		if updates > maxUpdates {
			maxUpdates = updates
			mostActiveCategory = category
		}
	}

	return Stats{
		TotalTasks:          len(tasks),
		LongestTaskId:       longestTask.taskId,
		AverageTaskDuration: totalDuration / len(tasks),
		MedianTaskDuration:  medianTask.duration,
		MostActiveTaskId:    mostActiveTask.taskId,
		MostActiveCategory:  mostActiveCategory,
	}
}

func GetTaskInfoBetween(tasks []Task, endDate time.Time, minDays int) (TaskInfo, error) {
	filteredTasks, err := getTasksByMinDays(tasks, minDays)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("filter tasks by min days: %w", err)
	}

	taskStats := getTaskStats(filteredTasks, endDate)

	return TaskInfo{
		Stats:     taskStats,
		Tasks: filteredTasks,
	}, nil
}

func GetFinishedTaskInfoBetween(tasks []Task, endDate time.Time, minDays int) (TaskInfo, error) {
	filteredTasks := getFinishedTasks(tasks)

	filteredTasks, err := getTasksByMinDays(filteredTasks, minDays)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("filter tasks by min days: %w", err)
	}

	taskStats := getTaskStats(filteredTasks, endDate)

	return TaskInfo{
		Stats:     taskStats,
		Tasks: filteredTasks,
	}, nil
}

func GetAbandonedTaskInfoBetween(tasks []Task, endDate time.Time, minDays int) (TaskInfo, error) {
	filteredTasks := getAbandonedTasks(tasks)

	filteredTasks, err := getTasksByMinDays(filteredTasks, minDays)
	if err != nil {
		return TaskInfo{}, fmt.Errorf("filter tasks by min days: %w", err)
	}

	taskStats := getTaskStats(filteredTasks, endDate)

	return TaskInfo{
		Stats:     taskStats,
		Tasks: filteredTasks,
	}, nil
}
