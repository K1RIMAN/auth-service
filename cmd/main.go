package main

import (
	"auth-service/internal/api"
	"auth-service/internal/config"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title Auth Service API
// @version 1.0
// @description Сервис аутентификации с использованием JWT токенов.
// @contact.name API Support
// @license.name MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	repo, err := repository.NewPostgresRepository(cfg.Database.GetConnectionString())
	if err != nil {
		log.Fatalf("Ошибка создания репозитория: %v", err)
	}
	defer repo.Close()

	authService := service.NewAuthService(repo, cfg)
	authMiddleware := middleware.NewAuthMiddleware(authService)
	authHandler := api.NewAuthHandler(authService)

	server := api.NewServer(cfg.Server.Port, authHandler, authMiddleware)

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	log.Printf("Сервер запущен на порту %s", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка остановки сервера: %v", err)
	}
}
