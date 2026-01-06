package user

import "errors"

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
)

type Repo interface {
	Create(u *User) error
	GetByEmail(email string) (*User, error)
}
