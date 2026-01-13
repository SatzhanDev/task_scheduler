package task

import (
	"errors"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("task not found")
)

type Service interface {
	Create(userID int, title string, dueAt *time.Time) (*Task, error)
	Get(userID, id int) (*Task, error)
	List(userID, limit, offset int) ([]Task, int, int, error)
}

type TaskService struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &TaskService{
		repo: repo,
	}
}

func (s *TaskService) Create(userID int, title string, dueAt *time.Time) (*Task, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	now := time.Now().UTC()
	task := &Task{
		UserID:    userID,
		Title:     title,
		DueAt:     dueAt,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Get(userID int, id int) (*Task, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	task, err := s.repo.Get(userID, id)
	if err != nil {
		return nil, err
	}
	return task, nil

}

func (s *TaskService) List(userID int, limit, offset int) ([]Task, int, int, error) {
	// 1. Валидация offset
	if userID <= 0 {
		return nil, 0, 0, ErrInvalidInput
	}

	if offset < 0 {
		return nil, 0, 0, ErrInvalidInput
	}

	// 2. Значения по умолчанию и ограничения
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	tasks, total, err := s.repo.List(userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	return tasks, total, limit, nil

}
