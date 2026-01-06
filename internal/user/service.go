package user

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrAuthFailed   = errors.New("invalid email or password")
)

type Service interface {
	Register(email, password string) (*User, error)
	Authenticate(email, password string) (*User, error)
}

type userService struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &userService{repo: repo}
}

func (s *userService) Register(email, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, ErrInvalidInput
	}

	// минимальная защита: чтобы не регистрировали "123"
	if len(password) < 6 {
		return nil, ErrInvalidInput
	}

	// 1) хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 2) создаём юзера
	u := &User{
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	// 3) сохраняем
	if err := s.repo.Create(u); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	// 4) безопасность: наружу пароль-хеш не отдаём
	u.PasswordHash = ""
	return u, nil
}

func (s *userService) Authenticate(email, password string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, ErrInvalidInput
	}
	// 1) ищем пользователя
	u, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrAuthFailed
		}
		return nil, err
	}
	// 2) сравниваем пароль с хешем
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrAuthFailed
	}
	// 3) наружу хеш не отдаём
	u.PasswordHash = ""
	return u, nil
}
