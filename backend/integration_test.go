package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGetAllCoursesIntegration проверяет интеграцию GetAllCourses.
func TestGetAllCoursesIntegration(t *testing.T) {
	// Создание нового HTTP-запроса для конечной точки GetAllCourses
	request, err := http.NewRequest("GET", "/all", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создание записи ответа для записи ответа
	response := httptest.NewRecorder()

	// Создание нового маршрутизатора Gin
	router := gin.Default()

	// Добавление обработчика GetAllCourses к маршрутизатору
	router.GET("/all", GetAllCourses)

	// Обслуживание HTTP-запроса с маршрутизатором
	router.ServeHTTP(response, request)

	// Проверка статуса кода ответа
	if response.Code != http.StatusOK {
		t.Errorf("Неправильный код состояния. Ожидалось: %d, Получено: %d", http.StatusOK, response.Code)
	}

	// Определение структуры для декодирования JSON-ответа
	var courses []Course

	// Декодирование тела ответа в определенную структуру
	err = json.Unmarshal(response.Body.Bytes(), &courses)
	if err != nil {
		t.Fatal(err)
	}

	// Проверка, что слайс курсов не пуст
	if len(courses) == 0 {
		t.Error("Курсы не найдены в ответе")
	}

	expectedName := "JavaScript Fundamentals"
	if courses[0].Name != expectedName {
		t.Errorf("Неправильное название курса. Ожидалось: %s, Получено: %s", expectedName, courses[0].Name)
	}
}
