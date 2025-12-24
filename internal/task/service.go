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
	List(limit, offset int) ([]Task, int, error)
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

func (s *TaskService) List(limit, offset int) ([]Task, int, error) {
	// 1. Валидация offset

	if offset < 0 {
		return nil, 0, ErrInvalidInput
	}

	// 2. Значения по умолчанию и ограничения
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 3. Блокируем доступ к данным
	s.mu.Lock()
	defer s.mu.Unlock()

	// 4. Сколько всего задач
	total := len(s.tasks)

	// 5. Если offset больше total — просто вернём пустой список
	if offset >= total {
		return []Task{}, total, nil
	}

	// 6. Преобразуем map -> slice
	all := make([]Task, 0, total)
	for _, t := range s.tasks {
		all = append(all, *t)
	}

	// 7. Вычисляем границы
	end := offset + limit
	if end > total {
		end = total
	}

	// 8. Возвращаем нужный кусок
	return all[offset:end], total, nil
}

func NewService() Service {
	return &TaskService{
		nextID: 1,
		tasks:  make(map[int]*Task),
	}
}
