package service

import (
	"auth-service/internal/models"

	"github.com/google/uuid"
)

// Service интерфейс бизнес-логики приложения
type Service interface {
	// Login создает новую сессию для пользователя и возвращает токены
	Login(userID uuid.UUID, userAgent, clientIP string) (*models.TokenPair, error)

	// Refresh обновляет пару токенов
	Refresh(refreshToken, userAgent, clientIP string) (*models.TokenPair, error)

	// Validate проверяет access токен и возвращает ID пользователя
	Validate(accessToken string) (uuid.UUID, error)

	// Logout деавторизует пользователя (делает токены недействительными)
	Logout(accessToken string) error
}
