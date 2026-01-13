package task

type Repo interface {
	Create(t *Task) error
	Get(userID, id int) (*Task, error)
	List(userID, limit, offset int) ([]Task, int, error)
}
