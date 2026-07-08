package ports

import (
	"knotwork/internal/todo/domain"
)

type Repository interface {
	GetTaskInfoBetween(startDate, endDate string, minDays int) (domain.TaskInfo, error)
	GetFinishedTaskInfoBetween(startDate, endDate string, minDays int) (domain.TaskInfo, error)
	GetAbandonedTaskInfoBetween(startDate, endDate string, minDays int) (domain.TaskInfo, error)
}
