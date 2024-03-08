package main

import (
	"gorm.io/gorm"
	"time"
)

// Модель пользователя
type User struct {
	gorm.Model
	Username string `json:"username"`
	Email    string `gorm:"unique"`
	Password string
	Role     string
	Courses  []*Course `gorm:"many2many:user_courses"`
}

// Модель курса
type Course struct {
	gorm.Model
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Duration    string  `json:"duration"`
	Price       string  `json:"price"`
	Places      string  `json:"places"`
	Category    string  `json:"category"`
	Users       []*User `gorm:"many2many:user_courses"`
}

// UserCourse является промежуточной моделью для связи между пользователями и курсами.
type UserCourse struct {
	UserID   uint
	CourseID uint
}

// Subscribers представляет модель списка подписчиков.
type Subscribers struct {
	UserId uint
}

// VerificationCode представляет модель верификационного кода, используемого для подтверждения электронной почты.
type VerificationCode struct {
	ID             uint      `gorm:"primaryKey"`
	Email          string    `gorm:"uniqueIndex"`
	Code           string    `gorm:"not null"`
	ExpirationTime time.Time `gorm:"not null"`
}
