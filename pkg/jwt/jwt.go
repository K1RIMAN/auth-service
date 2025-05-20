package jwt

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenClaims структура данных для JWT токена
type TokenClaims struct {
	UserID         string `json:"user_id"`
	RefreshTokenID string `json:"refresh_token_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateAccessToken создает JWT access token
func GenerateAccessToken(userID uuid.UUID, secret string, expiry time.Duration) (string, error) {
	claims := TokenClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	
	// Создаем подпись с использованием SHA512
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("ошибка подписи токена: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken создает случайный refresh token и его идентификатор
func GenerateRefreshToken() (string, string) {
	refreshToken := uuid.New().String()
	refreshTokenID := uuid.New().String()
	
	return refreshToken, refreshTokenID
}

// ValidateAccessToken проверяет валидность access токена
func ValidateAccessToken(tokenString, secret string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что алгоритм подписи токена - HS512
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != "HS512" {
			return nil, fmt.Errorf("неожиданный алгоритм подписи: %v", token.Header["alg"])
		}
		
		// Возвращаем секретный ключ для проверки подписи
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("невалидный токен")
	}

	return claims, nil
}

// HashRefreshToken создает bcrypt хеш refresh токена
func HashRefreshToken(refreshToken string) string {
	// Используем SHA-512 для хеширования refresh токена
	hash := sha512.Sum512([]byte(refreshToken))
	return fmt.Sprintf("%x", hash)
} 