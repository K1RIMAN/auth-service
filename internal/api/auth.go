package api

import (
	"auth-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler обработчик API запросов для авторизации
type AuthHandler struct {
	service service.Service
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(service service.Service) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// @Summary Получение токенов пользователя
// @Description Получение пары токенов (access и refresh) для указанного ID пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param user_id query string true "ID пользователя (GUID)"
// @Success 200 {object} models.TokenPair "Пара токенов"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// Получаем ID пользователя из параметров запроса
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":        "error",
			"error_code":    "INVALID_REQUEST",
			"error_message": "отсутствует параметр user_id",
		})
		return
	}

	// Парсим ID пользователя
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":        "error",
			"error_code":    "INVALID_USER_ID",
			"error_message": "некорректный формат ID пользователя",
		})
		return
	}

	// Получаем User-Agent и IP-адрес клиента
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	// Генерируем токены
	tokens, err := h.service.Login(userID, userAgent, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":        "error",
			"error_code":    "INTERNAL_ERROR",
			"error_message": "ошибка при генерации токенов",
		})
		return
	}

	// Возвращаем токены клиенту
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tokens,
	})
}

// @Summary Обновление токенов
// @Description Обновление пары токенов (access и refresh) с использованием refresh токена
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh токен (в формате base64)"
// @Success 200 {object} models.TokenPair "Новая пара токенов"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Невалидный refresh токен"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	// Получаем refresh токен из тела запроса
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":        "error",
			"error_code":    "INVALID_REQUEST",
			"error_message": "отсутствует параметр refresh_token",
		})
		return
	}

	// Получаем User-Agent и IP-адрес клиента
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	// Обновляем токены
	tokens, err := h.service.Refresh(request.RefreshToken, userAgent, clientIP)
	if err != nil {
		// Если ошибка связана с изменением User-Agent
		if err.Error() == "обновление токенов с другого устройства запрещено" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":        "error",
				"error_code":    "INVALID_USER_AGENT",
				"error_message": "обновление токенов с другого устройства запрещено",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"status":        "error",
			"error_code":    "INVALID_REFRESH_TOKEN",
			"error_message": "невалидный refresh токен",
		})
		return
	}

	// Возвращаем новые токены клиенту
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tokens,
	})
}

// @Summary Получение ID текущего пользователя
// @Description Получение ID пользователя, которому принадлежит текущий access токен
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse "ID пользователя"
// @Failure 401 {object} models.ErrorResponse "Не авторизован"
// @Router /user/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Получаем ID пользователя из контекста запроса
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":        "error",
			"error_code":    "UNAUTHORIZED",
			"error_message": "пользователь не авторизован",
		})
		return
	}

	// Возвращаем ID пользователя
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user_id": userID.(uuid.UUID).String(),
		},
	})
}

// @Summary Деавторизация пользователя
// @Description Деавторизует пользователя, после чего его токены становятся недействительными
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Response "Успешная деавторизация"
// @Failure 401 {object} models.ErrorResponse "Не авторизован"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Получаем access токен из контекста запроса
	accessToken, exists := c.Get("accessToken")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":        "error",
			"error_code":    "UNAUTHORIZED",
			"error_message": "пользователь не авторизован",
		})
		return
	}

	// Деавторизуем пользователя
	err := h.service.Logout(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":        "error",
			"error_code":    "INTERNAL_ERROR",
			"error_message": "ошибка при деавторизации пользователя",
		})
		return
	}

	// Возвращаем успешный статус
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "пользователь успешно деавторизован",
	})
}
