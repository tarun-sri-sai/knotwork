package ports

import (
	"time"

	"knotwork-core/internal/domain"
)

type Repository interface {
	ParseDate(dateStr string) (time.Time, error)
	GetTasksBetween(startDateStr, endDateStr string) ([]domain.Task, error)
}
