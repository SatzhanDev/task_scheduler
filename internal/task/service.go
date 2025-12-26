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
	Create(title string, dueAt *time.Time) (*Task, error)
	Get(id int) (*Task, error)
	List(limit, offset int) ([]Task, int, int, error)
}

type TaskService struct {
	// mu     sync.Mutex
	// nextID int
	// tasks  map[int]*Task
	repo Repo
}

func NewService(repo Repo) Service {
	return &TaskService{
		repo: repo,
	}
}

func (s *TaskService) Create(title string, dueAt *time.Time) (*Task, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}
	// s.mu.Lock()
	// defer s.mu.Unlock()

	// id := s.nextID
	// s.nextID++
	now := time.Now().UTC()
	task := &Task{
		Title:     title,
		DueAt:     dueAt,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	// s.tasks[id] = task
	if err := s.repo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Get(id int) (*Task, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	// s.mu.Lock()
	// defer s.mu.Unlock()
	// task, ok := s.tasks[id]
	// if !ok {
	// 	return nil, ErrNotFound
	// }
	task, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	return task, nil

}

func (s *TaskService) List(limit, offset int) ([]Task, int, int, error) {
	// 1. Валидация offset

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

	tasks, total, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	return tasks, total, limit, nil

	// 3. Блокируем доступ к данным
	// s.mu.Lock()
	// defer s.mu.Unlock()

	// 4. Сколько всего задач
	// total := len(s.tasks)

	// 5. Если offset больше total — просто вернём пустой список
	// if offset >= total {
	// 	return []Task{}, total, limit, nil
	// }

	// // 6. Преобразуем map -> slice
	// all := make([]Task, 0, total)
	// for _, t := range s.tasks {
	// 	all = append(all, *t)
	// }

	// //7. Сортировка. Бывает только при in memory памяти когда используем map.
	// // При SQL database так не будет
	// sort.Slice(all, func(i, j int) bool {
	// 	return all[i].ID < all[j].ID
	// })
	// // 8. Вычисляем границы
	// end := offset + limit
	// if end > total {
	// 	end = total
	// }

	// 9. Возвращаем нужный кусок
	// return all[offset:end], total, limit, nil
}
