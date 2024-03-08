package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func register(c *gin.Context) {
	// Начало процесса регистрации пользователя
	GetLogger().Info("Starting user registration")

	// Попытка привязать JSON-данные запроса к структуре User
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		// Обработка ошибки невалидного запроса
		GetLogger().Error("Invalid registration request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Проверка, существует ли уже аккаунт с указанным email
	if err := db.Where("email = ?", newUser.Email).First(&newUser).Error; err == nil {
		// Обработка случая, когда аккаунт уже существует
		fmt.Println(err)
		GetLogger().Error("Account already registered for email:", newUser.Email)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "The account is already registered"})
		return
	}

	// Определение роли пользователя на основе email и пароля
	if newUser.Password == "123" && newUser.Email == "tilesjan2005@gmail.com" {
		newUser.Role = "ADMIN"
	} else {
		newUser.Role = "USER"
	}

	// Хэширование пароля перед сохранением в базе данных
	fmt.Println(newUser.Password)
	hashedPassword, _ := HashPassword(newUser.Password)
	newUser.Password = hashedPassword
	fmt.Println(newUser.Password)

	// Создание записи пользователя в базе данных
	db.Create(&newUser)

	// Создание JWT-токена для пользователя
	signedToken, _ := CreateToken(strconv.Itoa(int(newUser.ID)), newUser.Email, newUser.Role)

	// Установка куки с JWT-токеном для пользователя
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    signedToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, &cookie)

	// Логирование успешной регистрации пользователя
	GetLogger().Info("User registered successfully")

	// Отправка ответа об успешной регистрации
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
func login(c *gin.Context) {
	// Начало процесса входа пользователя в систему
	GetLogger().Info("Starting user login")

	// Создание структуры для хранения данных формы входа
	var user User
	type loginForm struct {
		Username string `json:"username"`
		Email    string `json:"email,omitempty"`
		Password string `json:"password"`
	}

	// Привязка JSON-данных запроса к структуре формы входа
	var newUser loginForm
	if err := c.ShouldBindJSON(&newUser); err != nil {
		// Обработка ошибки невалидного запроса
		GetLogger().Error("Invalid registration request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(newUser)

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Поиск пользователя по email в базе данных
	if err := db.Where("email = ?", newUser.Email).First(&user).Error; err != nil {
		// Обработка ошибки неверного email или пароля
		GetLogger().Error("Invalid email or password", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	fmt.Println(user)
	fmt.Println(HashPassword(newUser.Password))
	fmt.Println((user.Password))

	// Проверка соответствия хэшированного пароля
	if !CheckPasswordHash(newUser.Password, user.Password) {
		// Обработка ошибки аутентификации
		GetLogger().Error("Authentication failed for user:", user.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	// Определение роли пользователя на основе email и пароля
	if user.Password == "qwerty123" && user.Email == "musabecova05@gmail.com" {
		user.Role = "ADMIN"
	} else {
		user.Role = "USER"
	}

	// Создание JWT-токена для пользователя
	token, err := CreateToken(strconv.Itoa(int(user.ID)), user.Email, user.Role)
	if err != nil {
		// Обработка ошибки создания токена
		GetLogger().Error("Failed to create token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token", "data": token})
		return
	}

	// Установка куки с JWT-токеном для пользователя
	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, &cookie)

	// Логирование успешного входа пользователя
	GetLogger().Info("User login successful")

	// Отправка ответа об успешном входе и передача токена
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func Getvcode(context *gin.Context) {
	// Начало процесса восстановления пароля
	GetLogger().Info("Starting forgot password process")

	// Структура для хранения данных формы запроса на восстановление пароля
	type form struct {
		Email string `json:"email"`
	}

	// Привязка JSON-данных запроса к структуре формы
	var forminput form
	if err := context.BindJSON(&forminput); err != nil {
		// Обработка невалидного запроса на восстановление пароля
		GetLogger().Error("Invalid forgot password request:", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	fmt.Println(forminput)

	// Генерация кода подтверждения
	verificationCode := GenerateVerificationCode()

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		GetLogger().Error("Failed to connect to the database:", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Сохранение кода подтверждения в базе данных
	err = SaveVerificationCode(context, db, forminput.Email, verificationCode)
	if err != nil {
		// Обработка ошибки сохранения кода подтверждения в базе данных
		GetLogger().Error("Failed to save verification code to Redis:", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Отправка кода подтверждения на электронную почту
	err = SendVerificationCodeEmail(forminput.Email, verificationCode)

	// Логирование успешной отправки кода подтверждения на электронную почту
	GetLogger().Info("Verification code sent to email successfully")

	// Отправка ответа об успешной отправке кода подтверждения на электронную почту
	context.JSON(http.StatusOK, gin.H{"message": "Verification code sent to your email"})
}
func Checkvcode(c *gin.Context) {
	// Проверка кода подтверждения
	GetLogger().Info("Checking verification code")

	// Структура для проверки кода подтверждения
	type CheckCode struct {
		Email string `json:"email,omitempty"`
		Code  string `json:"code"`
	}

	// Привязка JSON-данных запроса к структуре для проверки кода
	var code CheckCode
	if err := c.BindJSON(&code); err != nil {
		// Обработка невалидного запроса на проверку кода
		GetLogger().Error("Invalid code check request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Проверка кода подтверждения
	validCode, err := CheckVerificationCode(c, db, code.Email, code.Code)
	if err != nil {
		// Обработка ошибки проверки кода подтверждения
		GetLogger().Error("Failed to verify code:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify code"})
		return
	}

	// Если код не действителен, возврат ошибки
	if !validCode {
		GetLogger().Error("Failed to save user information")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user information"})
		return
	}

	// Успешная проверка кода
	GetLogger().Info("successful")
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}
func DeleteCourse(c *gin.Context) {
	// Удаление курса
	GetLogger().Info("Starting course delete")

	// Получение идентификатора курса из параметров запроса
	courseID := c.Param("id")

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Удаление курса из базы данных
	if err := db.Where("ID = ?", courseID).Delete(&Course{}).Error; err != nil {
		// Обработка ошибки удаления курса
		GetLogger().Error("Failed to bind JSON data for deleting course:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Успешное удаление курса
	GetLogger().Info("Course deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}
func UpdateUser(c *gin.Context) {
	// Обновление данных пользователя
	GetLogger().Info("Starting user update")

	// Получение электронной почты пользователя из контекста
	userEmail, emailExists := c.Get("email")
	if !emailExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found"})
		return
	}

	// Преобразование электронной почты в строку
	email, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert email to string"})
		return
	}

	// Получение данных пользователя из запроса
	var updateUser User
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		GetLogger().Error("Invalid user update request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Поиск существующего пользователя в базе данных
	var existingUser User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err != nil {
		// Обработка ошибки поиска пользователя
		GetLogger().Error("Failed to find user:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Обновление полей пользователя, если они не пустые
	if updateUser.Username != "" {
		existingUser.Username = updateUser.Username
	}
	if updateUser.Email != "" {
		existingUser.Email = updateUser.Email
	}
	if updateUser.Password != "" {
		// Обновление пароля (возможно, стоит хешировать перед сохранением)
		existingUser.Password = updateUser.Password
	}
	// Обновление других полей по необходимости

	// Сохранение обновленной информации о пользователе в базе данных
	if err := db.Save(&existingUser).Error; err != nil {
		// Обработка ошибки обновления пользователя
		GetLogger().Error("Failed to update user:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Успешное обновление данных пользователя
	GetLogger().Info("User updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": existingUser})
}

func GetCourse(c *gin.Context) {
	// Получение информации о курсе
	GetLogger().Info("Starting course getting")

	// Получение идентификатора курса из параметров запроса
	courseID := c.Param("id")

	// Подключение к базе данных
	var existingCourse Course
	db, err := dbConnect()
	if err != nil {
		// Обработка ошибки подключения к базе данных
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Поиск курса по идентификатору
	if err := db.First(&existingCourse, courseID).Error; err != nil {
		// Обработка ошибки поиска курса
		GetLogger().Error("Failed to find course:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Успешное получение информации о курсе
	GetLogger().Info("Course getting successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Course getting successfully", "course": existingCourse})
}
func Enroll(c *gin.Context) {
	// Получение информации о пользователе и курсе для записи на курс
	userEmail, emailExists := c.Get("email")
	if !emailExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found"})
		return
	}

	email, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert email to string"})
		return
	}

	// Получение идентификатора курса из параметров запроса
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Поиск пользователя
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		GetLogger().Error("Failed to find user:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		return
	}

	// Поиск курса
	var course Course
	if result := db.First(&course, courseID); result.Error != nil {
		GetLogger().Error("Failed to find course:", result.Error)
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Проверка, что пользователь еще не записан на курс
	for _, enrolledCourse := range user.Courses {
		if enrolledCourse.ID == uint(courseID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User is already enrolled in this course"})
			return
		}
	}

	// Создание записи о записи пользователя на курс
	userCourse := UserCourse{
		UserID:   user.ID,
		CourseID: course.ID,
	}
	db.Create(&userCourse)

	// Успешное завершение записи пользователя на курс
	c.JSON(http.StatusOK, gin.H{"message": "User enrolled in course successfully"})
}
func Profile(c *gin.Context) {
	// Получение электронной почты пользователя из контекста Gin
	userEmail, emailExists := c.Get("email")
	if !emailExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found"})
		return
	}

	// Проверка успешного преобразования электронной почты к строковому типу
	email, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert email to string"})
		return
	}

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Поиск пользователя в базе данных по электронной почте
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		fmt.Println(err)
		GetLogger().Error("Account already registered for email:", email)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "The account is already registered"})
		return
	}

	// Отправка ответа с информацией о пользователе
	c.JSON(http.StatusOK, gin.H{"message": "User enrolled in course successfully", "user": user})
}
func UpdateCourse(c *gin.Context) {
	// Логгирование начала обновления курса
	GetLogger().Info("Starting course update")
	// Получение идентификатора курса из параметров запроса
	courseID := c.Param("id")

	// Подключение к базе данных и поиск существующего курса по его идентификатору
	var existingCourse Course
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}
	if err := db.First(&existingCourse, courseID).Error; err != nil {
		GetLogger().Error("Failed to find course:", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}

	// Привязка новых данных о курсе из запроса
	var newCourse Course
	if err := c.ShouldBind(&newCourse); err != nil {
		GetLogger().Error("Failed to bind JSON data for updating course:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновление полей курса, если они указаны в запросе
	if newCourse.Name != "" {
		existingCourse.Name = newCourse.Name
	}
	if newCourse.Description != "" {
		existingCourse.Description = newCourse.Description
	}
	if newCourse.Image != "" {
		existingCourse.Image = newCourse.Image
	}
	if newCourse.Duration != "" {
		existingCourse.Duration = newCourse.Duration
	}
	if newCourse.Price != "" {
		existingCourse.Price = newCourse.Price
	}
	if newCourse.Places != "" {
		existingCourse.Places = newCourse.Places
	}
	if newCourse.Category != "" {
		existingCourse.Category = newCourse.Category
	}

	// Сохранение обновленных данных о курсе в базе данных
	if err := db.Save(&existingCourse).Error; err != nil {
		GetLogger().Error("Failed to update course:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update course"})
		return
	}

	// Отправка ответа об успешном обновлении курса
	GetLogger().Info("Course updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Course updated successfully", "course": existingCourse})
}
func CreateCourse(c *gin.Context) {
	// Логгирование начала создания курса
	GetLogger().Info("Starting course creation")

	// Привязка данных нового курса из запроса
	var newCourse Course
	if err := c.ShouldBind(&newCourse); err != nil {
		GetLogger().Error("Failed to bind JSON data for new course:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Создание курса в базе данных
	if err := db.Create(&newCourse).Error; err != nil {
		GetLogger().Error("Failed to create course:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course"})
		return
	}

	// Логгирование успешного создания курса
	GetLogger().Info("Course created successfully")

	// Отправка ответа об успешном создании курса
	c.JSON(http.StatusOK, gin.H{"message": "Course created successfully", "course": newCourse})
}
func GetCourses(c *gin.Context) {
	// Логгирование начала получения курсов
	GetLogger().Info("Starting to fetch courses")

	var courses []Course

	// Извлечение параметров запроса
	limitStr := c.DefaultQuery("limit", "9")
	pageStr := c.DefaultQuery("page", "1")
	sortBy := c.DefaultQuery("sort_by", "id")
	order := c.DefaultQuery("order", "asc")

	// Преобразование параметров запроса в числовые значения
	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)

	// Вычисление смещения для постраничного вывода
	offset := (page - 1) * limit

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Получение курсов с учетом параметров запроса
	result := db.Order(fmt.Sprintf("%s %s", sortBy, order)).Limit(limit).Offset(offset).Find(&courses)
	if result.Error != nil {
		GetLogger().Error("Failed to fetch courses:", result.Error.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses"})
		return
	}

	// Логгирование успешного получения курсов
	GetLogger().Info("Courses fetched successfully")

	// Отправка ответа с полученными курсами
	c.JSON(http.StatusOK, courses)
}
func GetAllCourses(c *gin.Context) {
	// Логгирование начала получения всех курсов
	GetLogger().Info("Starting to fetch all courses")

	var courses []Course

	// Подключение к базе данных
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Получение всех курсов из базы данных
	result := db.Find(&courses)
	if result.Error != nil {
		GetLogger().Error("Failed to fetch courses:", result.Error.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses"})
		return
	}

	// Логгирование успешного получения всех курсов
	GetLogger().Info("All courses fetched successfully")

	// Отправка ответа с полученными курсами
	c.JSON(http.StatusOK, courses)
}
func UserCourses(c *gin.Context) {
	// Извлекаем email пользователя из контекста
	userEmail, emailExists := c.Get("email")
	if !emailExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email пользователя не найден"})
		return
	}

	// Преобразуем email пользователя в строку
	email, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось преобразовать email в строку"})
		return
	}

	// Подключаемся к базе данных
	db, err := dbConnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подключения к базе данных"})
		return
	}

	// Находим пользователя в базе данных
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Получаем курсы пользователя
	var userCourses []Course
	if err := db.Model(&user).Association("Courses").Find(&userCourses); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить курсы пользователя"})
		return
	}

	// Возвращаем курсы пользователя
	c.JSON(http.StatusOK, gin.H{"courses": userCourses})
}
func Subscribe(c *gin.Context) {
	GetLogger().Info("Starting Subscribe user") // Начало процесса подписки пользователя

	userEmail, emailExists := c.Get("email") // Получаем email пользователя из контекста
	if !emailExists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found"}) // Если email пользователя не найден, возвращаем ошибку
		return
	}

	email, ok := userEmail.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert email to string"}) // Если не удалось преобразовать email в строку, возвращаем ошибку
		return
	}

	db, err := dbConnect() // Подключаемся к базе данных
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())                 // Если не удалось подключиться к базе данных, записываем ошибку в лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"}) // Возвращаем ошибку
		return
	}

	var count int64
	if err := db.Model(&Subscribers{}).Where("user_id = (SELECT id FROM users WHERE email = ?)", email).Count(&count).Error; err != nil {
		GetLogger().Error("Failed to check subscription:", err.Error())                        // Если не удалось проверить подписку, записываем ошибку в лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subscription"}) // Возвращаем ошибку
		return
	}

	// Если пользователь уже подписан, возвращаем ошибку
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already subscribed"})
		return
	}

	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		fmt.Println(err)
		GetLogger().Error("Account already registered for email:", email)                           // Если аккаунт уже зарегистрирован для данного email, записываем ошибку в лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "The account is already registered"}) // Возвращаем ошибку
		return
	}

	var subscribe Subscribers
	subscribe.UserId = user.ID
	if err := db.Create(&subscribe).Error; err != nil {
		GetLogger().Error("Failed to create subscribe:", err.Error())                     // Если не удалось создать подписку, записываем ошибку в лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course"}) // Возвращаем ошибку
		return
	}

	GetLogger().Info("subscribe  successfully") // Подписка успешно создана

	c.JSON(http.StatusOK, gin.H{"message": "subscribesuccessfully"}) // Возвращаем сообщение об успешной подписке
}
func SendSpam(c *gin.Context) {
	GetLogger().Info("Starting spam email sending") // Начало отправки спама по электронной почте

	type form struct {
		Text string `json:"text"`
	}
	var text form
	if err := c.ShouldBindJSON(&text); err != nil {
		GetLogger().Error("Invalid request:", err)                 // Если запрос недействителен, записываем ошибку в лог
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Возвращаем ошибку
		return
	}
	// Подключаемся к базе данных
	fmt.Println(text.Text)
	db, err := dbConnect()
	if err != nil {
		GetLogger().Error("Failed to connect to the database:", err.Error())                 // Если не удалось подключиться к базе данных, записываем ошибку в лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"}) // Возвращаем ошибку
		return
	}

	// Извлекаем всех подписчиков
	var subscribers []Subscribers
	if err := db.Find(&subscribers).Error; err != nil {
		GetLogger().Error("Failed to fetch subscribers:", err.Error())                        // Если не удалось получить подписчиков, записываем ошибку в лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"}) // Возвращаем ошибку
		return
	}

	// Получаем пользователей, соответствующих подписчикам
	var users []User
	for _, sub := range subscribers {
		var user User
		if err := db.Where("ID = ?", sub.UserId).First(&user).Error; err != nil {
			GetLogger().Error("Failed to find user for subscriber:", err.Error()) // Если не удалось найти пользователя для подписчика, записываем ошибку в лог
			continue
		}
		fmt.Println(user.Email)
		users = append(users, user)
	}

	// Отправляем уведомления по электронной почте пользователям
	for _, user := range users {
		_ = SendSpamForUser(user.Email, text.Text)
		GetLogger().Info("Spam email sent to:", user.Email) // Записываем в лог отправку спама на адрес пользователя
	}

	GetLogger().Info("Spam emails sent successfully")                        // Записываем в лог успешную отправку спама
	c.JSON(http.StatusOK, gin.H{"message": "Spam emails sent successfully"}) // Возвращаем сообщение об успешной отправке спама
}
