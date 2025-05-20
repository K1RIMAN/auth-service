package repository

import (
	"auth-service/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresRepository реализация Repository с использованием PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	// Проверяем соединение
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось проверить соединение с базой данных: %w", err)
	}

	// Создаем необходимые таблицы, если они не существуют
	if err = createTables(db); err != nil {
		return nil, fmt.Errorf("не удалось создать таблицы: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// createTables создает необходимые таблицы в базе данных
func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS sessions (
		id SERIAL PRIMARY KEY,
		user_id UUID NOT NULL,
		refresh_token TEXT NOT NULL,
		refresh_token_id TEXT NOT NULL,
		user_agent TEXT NOT NULL,
		client_ip TEXT NOT NULL,
		is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
		expires_at BIGINT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := db.Exec(query)
	return err
}

// CreateSession создает новую сессию пользователя
func (r *PostgresRepository) CreateSession(userID uuid.UUID, refreshToken, refreshTokenID, userAgent, clientIP string, expiresAt int64) (int, error) {
	var sessionID int
	query := `
	INSERT INTO sessions (user_id, refresh_token, refresh_token_id, user_agent, client_ip, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id
	`

	err := r.db.QueryRow(query, userID, refreshToken, refreshTokenID, userAgent, clientIP, expiresAt).Scan(&sessionID)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать сессию: %w", err)
	}

	return sessionID, nil
}

// GetSessionByRefreshToken возвращает сессию по хешу refresh токена
func (r *PostgresRepository) GetSessionByRefreshToken(refreshTokenHash string) (*models.Session, error) {
	query := `
	SELECT id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, refresh_token_id
	FROM sessions
	WHERE refresh_token = $1 AND is_blocked = FALSE AND expires_at > $2
	`

	row := r.db.QueryRow(query, refreshTokenHash, time.Now().Unix())
	
	session := &models.Session{}
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.ClientIP,
		&session.IsBlocked,
		&session.ExpiresAt,
		&session.RefreshTokenID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("сессия не найдена или истекла")
		}
		return nil, fmt.Errorf("ошибка получения сессии: %w", err)
	}

	return session, nil
}

// UpdateSession обновляет refresh токен в сессии
func (r *PostgresRepository) UpdateSession(sessionID int, refreshToken, refreshTokenID string, expiresAt int64) error {
	query := `
	UPDATE sessions
	SET refresh_token = $1, refresh_token_id = $2, expires_at = $3, updated_at = CURRENT_TIMESTAMP
	WHERE id = $4
	`

	_, err := r.db.Exec(query, refreshToken, refreshTokenID, expiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("не удалось обновить сессию: %w", err)
	}

	return nil
}

// BlockSession блокирует сессию
func (r *PostgresRepository) BlockSession(sessionID int) error {
	query := `
	UPDATE sessions
	SET is_blocked = TRUE, updated_at = CURRENT_TIMESTAMP
	WHERE id = $1
	`

	_, err := r.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("не удалось заблокировать сессию: %w", err)
	}

	return nil
}

// BlockAllUserSessions блокирует все сессии пользователя
func (r *PostgresRepository) BlockAllUserSessions(userID uuid.UUID) error {
	query := `
	UPDATE sessions
	SET is_blocked = TRUE, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = $1
	`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("не удалось заблокировать все сессии пользователя: %w", err)
	}

	return nil
}

// Close закрывает соединение с базой данных
func (r *PostgresRepository) Close() error {
	return r.db.Close()
} 