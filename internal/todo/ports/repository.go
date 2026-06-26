package ports

import (
	"knotwork-core/internal/todo/domain"
)

type Repository interface {
	GetTaskDurationsBetween(startDate, endDate string) ([]domain.TaskDuration, error)
}
