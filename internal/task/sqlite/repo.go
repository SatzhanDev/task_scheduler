package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"task_scheduler/internal/task"
	"time"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, t *task.Task) error {
	// 1) Готовим значения для due_at: либо NULL, либо строка
	var dueAt sql.NullString
	if t.DueAt != nil {
		dueAt = sql.NullString{String: t.DueAt.UTC().Format(time.RFC3339Nano), Valid: true}
	}

	// 2) Вставляем запись
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (user_id, title, due_at, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.UserID,
		t.Title,
		dueAt,
		string(t.Status),
		t.CreatedAt.UTC().Format(time.RFC3339Nano),
		t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return err
	}

	// 3) Забираем id, который сгенерировала БД
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = int(id)
	return nil
}

func (r *Repo) Get(ctx context.Context, userID, id int) (*task.Task, error) {
	var (
		t            task.Task
		dueAt        sql.NullString
		StatusStr    string
		createdAtStr string
		updatedAtStr string
	)

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, due_at, status, created_at, updated_at
		 FROM tasks
		 WHERE user_id = ? AND id = ?`,
		userID,
		id,
	).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&dueAt,
		&StatusStr,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, task.ErrNotFound
		}
		return nil, err
	}

	// due_at может быть NULL
	if dueAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, dueAt.String)
		if err != nil {
			return nil, err
		}
		t.DueAt = &parsed
	}

	// created_at / updated_at парсим из строк
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, err
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return nil, err
	}
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	// status в модели — Status (string alias)
	t.Status = task.Status(StatusStr)
	return &t, nil
}

func (r *Repo) List(ctx context.Context, userID, limit, offset int) ([]task.Task, int, error) {
	// 1) Total
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2) Page rows
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, title, due_at, status, created_at, updated_at
		 FROM tasks
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]task.Task, 0)

	for rows.Next() {
		var (
			t          task.Task
			dueAt      sql.NullString
			createdStr string
			updatedStr string
			statusStr  string
		)

		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&dueAt,
			&statusStr,
			&createdStr,
			&updatedStr,
		); err != nil {
			return nil, 0, err
		}

		// status в модели — Status (string alias)
		t.Status = task.Status(statusStr)

		if dueAt.Valid {
			parsed, err := time.Parse(time.RFC3339Nano, dueAt.String)
			if err != nil {
				return nil, 0, err
			}
			t.DueAt = &parsed
		}

		createdAt, err := time.Parse(time.RFC3339Nano, createdStr)
		if err != nil {
			return nil, 0, err
		}
		updatedAt, err := time.Parse(time.RFC3339Nano, updatedStr)
		if err != nil {
			return nil, 0, err
		}
		t.CreatedAt = createdAt
		t.UpdatedAt = updatedAt

		tasks = append(tasks, t)
	}

	// 3) rows.Err — ошибки итерации
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *Repo) Update(ctx context.Context, t *task.Task) error {
	var dueAt sql.NullString
	if t.DueAt != nil {
		dueAt = sql.NullString{String: t.DueAt.UTC().Format(time.RFC3339Nano)}
	}

	res, err := r.db.ExecContext(ctx, `UPDATE tasks SET title = ?, due_at = ?, status = ?, updated_at = ? WHERE user_id = ? AND id = ?`, t.Title, dueAt, string(t.Status), t.UpdatedAt.UTC().Format(time.RFC3339Nano), t.UserID, t.ID)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return task.ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, userID, id int) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE user_id = ? AND od = ?`,
		userID,
		id,
	)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return task.ErrNotFound
	}
	return nil
}
