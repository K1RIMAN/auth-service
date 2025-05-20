package api

import (
	"auth-service/internal/middleware"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Server представляет HTTP сервер
type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

// NewServer создает новый экземпляр сервера
func NewServer(port string, handler *AuthHandler, authMiddleware *middleware.AuthMiddleware) *Server {
	// Создаем роутер
	router := gin.Default()

	// Настраиваем middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Добавляем Swagger документацию
	router.GET("/swagger/*any", gin.WrapH(http.StripPrefix("/swagger/", http.FileServer(http.Dir("./swagger")))))

	// Группа роутов для авторизации
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.Refresh)
		authGroup.POST("/logout", authMiddleware.CheckAuth(), handler.Logout)
	}

	// Группа роутов для пользователя
	userGroup := router.Group("/user")
	{
		userGroup.GET("/me", authMiddleware.CheckAuth(), handler.GetCurrentUser)
	}

	// Создаем HTTP сервер
	httpServer := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return &Server{
		httpServer: httpServer,
		router:     router,
	}
}

// Run запускает HTTP сервер
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown останавливает HTTP сервер
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// GetRouter возвращает роутер для тестирования
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
