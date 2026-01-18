package handlers

import (
	"net/http"
	"strconv"

	"coalyonysh/go-musthave-diploma-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	storage storage.Storage
}

func NewBalanceHandler(storage storage.Storage) *BalanceHandler {
	return &BalanceHandler{storage: storage}
}

func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.storage.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current":   user.Balance,
		"withdrawn": user.Withdrawn,
	})
}

func (h *BalanceHandler) Withdraw(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order or sum"})
		return
	}

	if !h.isValidLuhn(req.Order) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		return
	}

	// Получить пользователя
	user, err := h.storage.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	if user.Balance < req.Sum {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient funds"})
		return
	}

	// Создать withdrawal
	err = h.storage.CreateWithdrawal(userID.(int), req.Order, req.Sum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Обновить баланс
	newBalance := user.Balance - req.Sum
	newWithdrawn := user.Withdrawn + req.Sum
	err = h.storage.UpdateUserBalance(userID.(int), newBalance, newWithdrawn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "withdrawal successful"})
}

func (h *BalanceHandler) GetWithdrawals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	withdrawals, err := h.storage.GetWithdrawalsByUserID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(withdrawals) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}

func (h *BalanceHandler) isValidLuhn(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
