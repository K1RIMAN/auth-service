package middleware

import (
	"auth-service/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// AuthMiddleware middleware для проверки авторизации
type AuthMiddleware struct {
	service service.Service
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(service service.Service) *AuthMiddleware {
	return &AuthMiddleware{
		service: service,
	}
}

// CheckAuth проверяет валидность access токена
func (m *AuthMiddleware) CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "отсутствует заголовок Authorization",
			})
			c.Abort()
			return
		}

		// Проверяем формат заголовка
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "неверный формат заголовка Authorization",
			})
			c.Abort()
			return
		}

		// Получаем токен
		tokenString := parts[1]

		// Проверяем токен
		userID, err := m.service.Validate(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "невалидный токен",
			})
			c.Abort()
			return
		}

		// Сохраняем ID пользователя в контексте запроса
		c.Set("userID", userID)
		c.Set("accessToken", tokenString)

		c.Next()
	}
} 