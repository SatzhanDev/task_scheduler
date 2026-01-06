package sqlite

import (
	"database/sql"
	"errors"
	"strings"
	"task_scheduler/internal/user"
	"time"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(u *user.User) error {
	query := `INSERT INTO users(email, password_hash, created_at) VALUES (?, ?, ?)`

	res, err := r.db.Exec(
		query,
		u.Email,
		u.PasswordHash,
		u.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return user.ErrEmailExists
		}
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)
	return nil
}

func isUniqueViolation(err error) bool {
	var sqliteErr interface {
		Error() string
	}
	if errors.As(err, &sqliteErr) {
		return strings.Contains(err.Error(), "UNIQUE")
	}
	return false
}

func (r *Repo) GetByEmail(email string) (*user.User, error) {
	query := `
	SELECT id, email, password_hash, created_at
	FROM users
	WHERE email = ?
	`
	var u user.User
	var createdAtStr string

	err := r.db.QueryRow(query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&createdAtStr,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	t, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, err
	}
	u.CreatedAt = t
	return &u, nil
}
