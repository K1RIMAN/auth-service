package repository

import (
	"auth-service/internal/models"

	"github.com/google/uuid"
)

// Repository интерфейс для работы с данными
type Repository interface {
	// CreateSession создает новую сессию для пользователя
	CreateSession(userID uuid.UUID, refreshToken, refreshTokenID, userAgent, clientIP string, expiresAt int64) (int, error)

	// GetSessionByRefreshToken получает сессию по refresh токену
	GetSessionByRefreshToken(refreshTokenHash string) (*models.Session, error)

	// UpdateSession обновляет сессию
	UpdateSession(sessionID int, refreshToken, refreshTokenID string, expiresAt int64) error

	// BlockSession блокирует сессию
	BlockSession(sessionID int) error

	// BlockAllUserSessions блокирует все сессии пользователя
	BlockAllUserSessions(userID uuid.UUID) error

	// Close закрывает соединение с базой данных
	Close() error
}
