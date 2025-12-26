package task

type Repo interface {
	Create(t *Task) error
	Get(id int) (*Task, error)
	List(limit, offset int) ([]Task, int, int, error)
}
