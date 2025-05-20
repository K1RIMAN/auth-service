package models

import "github.com/google/uuid"

// User представляет модель пользователя в системе
type User struct {
	ID uuid.UUID `json:"id" db:"id"`
}

// Session представляет сессию пользователя
type Session struct {
	ID            int       `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	RefreshToken  string    `json:"-" db:"refresh_token"`
	UserAgent     string    `json:"-" db:"user_agent"`
	ClientIP      string    `json:"-" db:"client_ip"`
	IsBlocked     bool      `json:"-" db:"is_blocked"`
	ExpiresAt     int64     `json:"-" db:"expires_at"`
	RefreshTokenID string    `json:"-" db:"refresh_token_id"`
} 