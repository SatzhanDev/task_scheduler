package task

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("task not found")
)

type Service interface {
	Create(ctx context.Context, userID int, title string, dueAt *time.Time) (*Task, error)
	Get(ctx context.Context, userID, id int) (*Task, error)
	List(ctx context.Context, userID, limit, offset int) ([]Task, int, int, error)
	Update(ctx context.Context, userId, id int, input UpdateTaskInput) (*Task, error)
	Delete(ctx context.Context, userID, id int) error
}

type TaskService struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &TaskService{
		repo: repo,
	}
}

func (s *TaskService) Create(ctx context.Context, userID int, title string, dueAt *time.Time) (*Task, error) {
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
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Get(ctx context.Context, userID int, id int) (*Task, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	task, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return task, nil

}

func (s *TaskService) List(ctx context.Context, userID int, limit, offset int) ([]Task, int, int, error) {
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

	tasks, total, err := s.repo.List(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	return tasks, total, limit, nil

}

func (s *TaskService) Update(ctx context.Context, userID int, id int, input UpdateTaskInput) (*Task, error) {
	if userID <= 0 || id <= 0 {
		return nil, ErrInvalidInput
	}
	// PATCH без полей — ошибка (на всякий, даже если handler уже проверяет)
	if input.Title == nil && input.Status == nil && !input.DueAt.Set {
		return nil, ErrInvalidInput
	}
	// 1) Берём текущую задачу (сразу проверка ownership)
	tsk, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	// 2) Title
	if input.Title != nil {
		if *input.Title == "" {
			return nil, ErrInvalidInput
		}
		tsk.Title = *input.Title
	}

	// 3) Status
	if input.Status != nil {
		switch *input.Status {
		case string(StatusPending), string(StatusDone), string(StatusCanceled):
			tsk.Status = Status(*input.Status)
		default:
			return nil, ErrInvalidInput
		}
	}

	// 4) DueAt (3 состояния)
	if input.DueAt.Set {
		if input.DueAt.Value == nil {
			tsk.DueAt = nil
		} else {
			tsk.DueAt = input.DueAt.Value
		}
	}
	tsk.UpdatedAt = time.Now().UTC()

	// 5) Сохраняем
	if err := s.repo.Update(ctx, tsk); err != nil {
		return nil, err
	}
	return tsk, nil

}

func (s *TaskService) Delete(ctx context.Context, userID, id int) error {
	if userID <= 0 || id <= 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, userID, id)
}
