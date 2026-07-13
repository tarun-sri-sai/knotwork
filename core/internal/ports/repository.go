package ports

import (
	"time"

	"knotwork-core/internal/domain"
)

type Repository interface {
	ParseDate(dateStr string) (time.Time, error)
	GetTaskDurationsBetween(startDateStr, endDateStr string) ([]domain.TaskDuration, error)
}
