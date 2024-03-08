package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateAndVerifyToken - функция для тестирования создания и проверки JWT токена
func TestCreateAndVerifyToken(t *testing.T) {
	// Установка переменной окружения JWT_SECRET для тестирования
	os.Setenv("JWT_SECRET", "test_secret")

	// Тестирование функции CreateToken
	id := "user123"
	email := "user@example.com"
	userType := "regular"
	token, err := CreateToken(id, email, userType)
	assert.NoError(t, err, "Ошибка при создании токена")

	// Тестирование функции VerifyToken
	parsedId, parsedEmail, parsedUserType, err := VerifyToken(token)
	assert.NoError(t, err, "Ошибка при проверке токена")
	assert.Equal(t, id, parsedId, "Ожидается, что идентификатор совпадает")
	assert.Equal(t, email, parsedEmail, "Ожидается, что email совпадает")
	assert.Equal(t, userType, parsedUserType, "Ожидается, что тип пользователя совпадает")

	// Тестирование недействительного токена
	invalidToken := "invalid_token"
	_, _, _, err = VerifyToken(invalidToken)
	assert.Error(t, err, "Ожидается ошибка для недействительного токена")
}

// TestVerifyTokenWithEmptyToken - функция для тестирования проверки пустого токена
func TestVerifyTokenWithEmptyToken(t *testing.T) {
	// Тестирование функции VerifyToken с пустым токеном
	id, email, userType, err := VerifyToken("")
	assert.Empty(t, id, "Ожидается пустой ID для пустого токена")
	assert.Empty(t, email, "Ожидается пустой email для пустого токена")
	assert.Empty(t, userType, "Ожидается пустой userType для пустого токена")
	assert.Error(t, err, "Ожидается ошибка для пустого токена")
}

// TestVerifyTokenWithInvalidSignature - функция для тестирования проверки токена с недействительной подписью
func TestVerifyTokenWithInvalidSignature(t *testing.T) {
	// Тестирование функции VerifyToken с токеном с недействительной подписью
	invalidToken := "invalid_signature_token"
	id, email, userType, err := VerifyToken(invalidToken)
	assert.Empty(t, id, "Ожидается пустой ID для токена с недействительной подписью")
	assert.Empty(t, email, "Ожидается пустой email для токена с недействительной подписью")
	assert.Empty(t, userType, "Ожидается пустой userType для токена с недействительной подписью")
	assert.Error(t, err, "Ожидается ошибка для токена с недействительной подписью")
}
