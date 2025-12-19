package task

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("task not found")
)

type Service interface {
	Create(title string, dueAt *time.Time) (*Task, error)
	Get(id int) (*Task, error)
}

type TaskService struct {
	mu     sync.Mutex
	nextID int
	tasks  map[int]*Task
}

func (s *TaskService) Create(title string, dueAt *time.Time) (*Task, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID
	s.nextID++
	createdAt := time.Now().UTC()
	task := &Task{
		ID:        id,
		Title:     title,
		DueAt:     dueAt,
		Status:    StatusPending,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	s.tasks[id] = task
	return task, nil
}

func (s *TaskService) Get(id int) (*Task, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return task, nil

}

func NewService() Service {
	return &TaskService{
		nextID: 1,
		tasks:  make(map[int]*Task),
	}
}
