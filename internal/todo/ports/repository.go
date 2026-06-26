package ports

import (
	"knotwork/internal/todo/domain"
)

type Repository interface {
	GetTaskDurationsBetween(startDate, endDate string) ([]domain.TaskDuration, error)
}
