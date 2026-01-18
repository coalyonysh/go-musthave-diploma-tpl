package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"coalyonysh/go-musthave-diploma-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

type OrdersHandler struct {
	storage storage.Storage
}

func NewOrdersHandler(storage storage.Storage) *OrdersHandler {
	return &OrdersHandler{storage: storage}
}

func (h *OrdersHandler) UploadOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order number required"})
		return
	}

	if !h.isValidLuhn(orderNumber) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		return
	}

	// Проверяем, существует ли заказ
	existingOrder, err := h.storage.GetOrderByNumber(orderNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID.(int) {
			c.JSON(http.StatusOK, gin.H{"message": "order already uploaded by this user"})
			return
		} else {
			c.JSON(http.StatusConflict, gin.H{"error": "order already uploaded by another user"})
			return
		}
	}

	// Создаем заказ
	_, err = h.storage.CreateOrder(userID.(int), orderNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "order uploaded successfully"})
}

func (h *OrdersHandler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	orders, err := h.storage.GetOrdersByUserID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrdersHandler) isValidLuhn(number string) bool {
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
