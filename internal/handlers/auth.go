package handlers

import (
	"net/http"
	"time"

	"coalyonysh/go-musthave-diploma-tpl/internal/models"
	"coalyonysh/go-musthave-diploma-tpl/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key")

type AuthHandler struct {
	storage storage.Storage
}

func NewAuthHandler(storage storage.Storage) *AuthHandler {
	return &AuthHandler{storage: storage}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Login == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login and password required"})
		return
	}

	// Проверяем, существует ли пользователь
	existing, err := h.storage.GetUserByLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "login already exists"})
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Создаем пользователя
	user, err := h.storage.CreateUser(req.Login, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Генерируем токен
	token, err := h.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Login == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "login and password required"})
		return
	}

	// Получаем пользователя
	user, err := h.storage.GetUserByLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		return
	}

	// Генерируем токен
	token, err := h.generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (h *AuthHandler) generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
