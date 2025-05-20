package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config структура содержит все конфигурационные параметры приложения
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Webhook  WebhookConfig
}

// ServerConfig содержит конфигурацию веб-сервера
type ServerConfig struct {
	Port string
}

// DatabaseConfig содержит конфигурацию подключения к базе данных
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig содержит конфигурацию для JWT токенов
type JWTConfig struct {
	AccessSecret  string
	AccessExpiry  time.Duration
	RefreshSecret string
	RefreshExpiry time.Duration
}

// WebhookConfig содержит конфигурацию для webhook
type WebhookConfig struct {
	URL string
}

// LoadConfig загружает конфигурацию из .env файла и переменных окружения
func LoadConfig() (*Config, error) {
	// Пытаемся загрузить .env файл, если он существует
	_ = godotenv.Load() // Игнорируем ошибку, если файл не существует

	cfg := &Config{}

	// Настройки сервера
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")

	// Настройки базы данных
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "postgres")
	cfg.Database.DBName = getEnv("DB_NAME", "auth_service_db")
	cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")

	// Настройки JWT
	cfg.JWT.AccessSecret = getEnv("JWT_ACCESS_SECRET", "default_access_secret")
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга JWT_ACCESS_EXPIRY: %w", err)
	}
	cfg.JWT.AccessExpiry = accessExpiry

	cfg.JWT.RefreshSecret = getEnv("JWT_REFRESH_SECRET", "default_refresh_secret")
	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "720h"))
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга JWT_REFRESH_EXPIRY: %w", err)
	}
	cfg.JWT.RefreshExpiry = refreshExpiry

	// Webhook URL
	cfg.Webhook.URL = getEnv("WEBHOOK_URL", "")

	return cfg, nil
}

// GetConnectionString возвращает строку подключения к PostgreSQL
func (dc *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.User, dc.Password, dc.DBName, dc.SSLMode)
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt получает численное значение переменной окружения
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
