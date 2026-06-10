package ports

import (
	"knotwork-core/internal/todo/domain"
)

type Repository interface {
	GetTaskMapBefore(date string) ([]domain.TaskMap, error)
	GetTaskMapBetween(startDate, endDate string) ([]domain.TaskMap, error)
}
