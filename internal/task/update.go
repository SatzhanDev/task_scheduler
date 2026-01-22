package task

import "time"

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

type UpdateTaskInput struct {
	Title  *string
	Status *string
	DueAt  OptionalTime
}
