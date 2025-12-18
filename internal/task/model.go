package task

import "time"

type Status string

const (
	StatusPending  Status = "pending"
	StatusDone     Status = "done"
	StatusCanceled Status = "canceled"
)

type Task struct {
	ID        int
	Title     string
	DueAt     *time.Time
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}
