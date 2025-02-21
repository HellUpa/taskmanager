package models

import (
	"time"
)

type Task struct {
	ID          int
	Title       string
	Description string
	DueDate     time.Time
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
