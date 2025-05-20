package service

import (
	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/pkg/jwt"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AuthService реализация сервиса авторизации
type AuthService struct {
	repo   repository.Repository
	config *config.Config
}

// LoginRequest структура для отправки webhook о попытке входа с нового IP
type LoginRequest struct {
	UserID  string `json:"user_id"`
	OldIP   string `json:"old_ip"`
	NewIP   string `json:"new_ip"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

// NewAuthService создает новый экземпляр сервиса авторизации
func NewAuthService(repo repository.Repository, config *config.Config) *AuthService {
	return &AuthService{
		repo:   repo,
		config: config,
	}
}

// Login создает новую сессию для пользователя и возвращает пару токенов
func (s *AuthService) Login(userID uuid.UUID, userAgent, clientIP string) (*models.TokenPair, error) {
	// Генерируем access токен
	accessToken, err := jwt.GenerateAccessToken(userID, s.config.JWT.AccessSecret, s.config.JWT.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания access токена: %w", err)
	}

	// Генерируем refresh токен и его ID
	refreshToken, refreshTokenID := jwt.GenerateRefreshToken()

	// Хешируем refresh токен для хранения в базе данных
	hashedRefreshToken := jwt.HashRefreshToken(refreshToken)

	// Вычисляем время истечения refresh токена
	expiresAt := time.Now().Add(s.config.JWT.RefreshExpiry).Unix()

	// Сохраняем сессию в базе данных
	_, err = s.repo.CreateSession(userID, hashedRefreshToken, refreshTokenID, userAgent, clientIP, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения сессии: %w", err)
	}

	// Кодируем refresh токен в base64 для передачи клиенту
	refreshTokenBase64 := base64.StdEncoding.EncodeToString([]byte(refreshToken))

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenBase64,
	}, nil
}

// Refresh обновляет пару токенов
func (s *AuthService) Refresh(refreshTokenBase64, userAgent, clientIP string) (*models.TokenPair, error) {
	// Декодируем refresh токен из base64
	refreshTokenBytes, err := base64.StdEncoding.DecodeString(refreshTokenBase64)
	if err != nil {
		return nil, fmt.Errorf("неверный формат refresh токена: %w", err)
	}
	refreshToken := string(refreshTokenBytes)

	// Хешируем refresh токен для поиска в базе данных
	hashedRefreshToken := jwt.HashRefreshToken(refreshToken)

	// Получаем сессию по refresh токену
	session, err := s.repo.GetSessionByRefreshToken(hashedRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена: %w", err)
	}

	// Проверяем, что User-Agent совпадает
	if session.UserAgent != userAgent {
		// Блокируем все сессии пользователя при попытке обновления токенов с другого устройства
		_ = s.repo.BlockAllUserSessions(session.UserID)
		return nil, errors.New("обновление токенов с другого устройства запрещено")
	}

	// Проверяем IP-адрес
	if session.ClientIP != clientIP {
		// Отправляем webhook о попытке входа с нового IP
		go s.sendLoginWebhook(session.UserID, session.ClientIP, clientIP)
	}

	// Генерируем новые токены
	accessToken, err := jwt.GenerateAccessToken(session.UserID, s.config.JWT.AccessSecret, s.config.JWT.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания access токена: %w", err)
	}

	// Генерируем новый refresh токен и его ID
	newRefreshToken, newRefreshTokenID := jwt.GenerateRefreshToken()

	// Хешируем новый refresh токен для хранения в базе данных
	hashedNewRefreshToken := jwt.HashRefreshToken(newRefreshToken)

	// Вычисляем время истечения нового refresh токена
	expiresAt := time.Now().Add(s.config.JWT.RefreshExpiry).Unix()

	// Обновляем сессию в базе данных
	err = s.repo.UpdateSession(session.ID, hashedNewRefreshToken, newRefreshTokenID, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления сессии: %w", err)
	}

	// Кодируем новый refresh токен в base64 для передачи клиенту
	newRefreshTokenBase64 := base64.StdEncoding.EncodeToString([]byte(newRefreshToken))

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenBase64,
	}, nil
}

// Validate проверяет access токен и возвращает ID пользователя
func (s *AuthService) Validate(accessToken string) (uuid.UUID, error) {
	// Проверяем валидность access токена
	claims, err := jwt.ValidateAccessToken(accessToken, s.config.JWT.AccessSecret)
	if err != nil {
		return uuid.Nil, fmt.Errorf("невалидный access токен: %w", err)
	}

	// Парсим ID пользователя
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("неверный формат ID пользователя: %w", err)
	}

	return userID, nil
}

// Logout деавторизует пользователя (делает токены недействительными)
func (s *AuthService) Logout(accessToken string) error {
	// Проверяем валидность access токена
	claims, err := jwt.ValidateAccessToken(accessToken, s.config.JWT.AccessSecret)
	if err != nil {
		return fmt.Errorf("невалидный access токен: %w", err)
	}

	// Парсим ID пользователя
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fmt.Errorf("неверный формат ID пользователя: %w", err)
	}

	// Блокируем все сессии пользователя
	err = s.repo.BlockAllUserSessions(userID)
	if err != nil {
		return fmt.Errorf("ошибка блокировки сессий: %w", err)
	}

	return nil
}

// sendLoginWebhook отправляет webhook о попытке входа с нового IP
func (s *AuthService) sendLoginWebhook(userID uuid.UUID, oldIP, newIP string) {
	// Проверяем, задан ли URL для webhook
	if s.config.Webhook.URL == "" {
		return
	}

	// Готовим данные для отправки
	data := LoginRequest{
		UserID:  userID.String(),
		OldIP:   oldIP,
		NewIP:   newIP,
		Time:    time.Now().Format(time.RFC3339),
		Message: "Обнаружена попытка обновления токенов с нового IP-адреса",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Ошибка сериализации данных для webhook: %v\n", err)
		return
	}

	// Отправляем запрос
	resp, err := http.Post(s.config.Webhook.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка отправки webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Printf("Webhook вернул ошибку: HTTP %d\n", resp.StatusCode)
	}
}
