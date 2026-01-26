package task

import "context"

type Repo interface {
	Create(ctx context.Context, t *Task) error
	Get(ctx context.Context, userID, id int) (*Task, error)
	List(ctx context.Context, userID, limit, offset int) ([]Task, int, error)
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, userID, id int) error
}
