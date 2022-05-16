package service

import (
	"crypto/sha1"
	"fmt"
	"todo-app"
	"todo-app/pkg/repository"
)

const salt = "hjqrhjqw124617ajfhajs" // "соль" для хэширования пароля

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(user todo.User) (int, error) {
	user.Password = generatePasswordHash(user.Password) // Перезаписываем пароль на хэшированный
	return s.repo.CreateUser(user)
}

func generatePasswordHash(password string) string { // Хэширование пароля
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
