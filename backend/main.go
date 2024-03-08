package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// rateLimiterMiddleware создает и возвращает middleware для ограничения скорости запросов.
// Он использует переданный лимитер, чтобы проверить, не превышает ли количество запросов
// установленное ограничение. Если превышение обнаружено, он возвращает ответ с кодом 429 Too Many Requests.
func rateLimiterMiddleware(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	// Выполнение автомиграции базы данных перед запуском сервера
	autoMigrate()

	// Создание лимитера скорости, разрешающего не более 3 запросов в секунду
	limiter := rate.NewLimiter(1, 3)

	// Создание нового экземпляра Gin с использованием настроек по умолчанию
	r := gin.Default()

	// Применение middleware для ограничения скорости запросов ко всем маршрутам
	r.Use(rateLimiterMiddleware(limiter))

	// Установка обработчиков маршрутов для различных эндпоинтов
	r.POST("/register", register)
	r.POST("/login", login)
	r.POST("/getvcode", Getvcode)
	r.POST("/checkvcode", Checkvcode)
	r.POST("/update", RequireAuthMiddleware, UpdateUser)
	r.GET("/profile", RequireAuthMiddleware, Profile)
	r.GET("/subscribe", RequireAuthMiddleware, Subscribe)
	r.POST("/admin/send", RequireAuthMiddleware, isAdmin, SendSpam)
	r.GET("/all", GetAllCourses)

	// Группирование маршрутов для управления курсами
	course := r.Group("/courses", RequireAuthMiddleware)
	{
		course.GET("/", GetCourses)
		course.GET("/:id", GetCourse)
		course.POST("/create", isAdmin, CreateCourse)
		course.PUT("/:id", isAdmin, UpdateCourse)
		course.DELETE("/:id", isAdmin, DeleteCourse)
		course.POST("/:id", Enroll)
		course.GET("/my", UserCourses)
	}

	// Создание HTTP-сервера с настройками адреса и обработчика маршрутов
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Запуск HTTP-сервера в горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe: %v\n", err)
		}
	}()

	// Ожидание сигнала прерывания (например, Ctrl+C) для завершения работы сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Shutting down server...")

	// Определение контекста с таймаутом для graceful shutdown сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Выполнение graceful shutdown сервера с ожиданием завершения всех запросов
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}
	fmt.Println("Server exiting")
}
