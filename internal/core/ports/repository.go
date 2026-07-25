package ports

import (
	"time"

	"github.com/tarun-sri-sai/knotwork/internal/core/domain"
)

type Repository interface {
	ParseDate(dateStr string) (time.Time, error)
	GetTasksBetween(startDateStr, endDateStr string) ([]domain.Task, error)
}
