package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"coalyonysh/go-musthave-diploma-tpl/internal/services"
	"coalyonysh/go-musthave-diploma-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

type OrdersHandler struct {
	storage    storage.Storage
	accrualSvc *services.AccrualService
}

func NewOrdersHandler(storage storage.Storage, accrualSvc *services.AccrualService) *OrdersHandler {
	return &OrdersHandler{storage: storage, accrualSvc: accrualSvc}
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

	// Обновляем статус сразу
	h.updateOrderStatuses()

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

func (h *OrdersHandler) updateOrderStatuses() {
	// Получить заказы в NEW и PROCESSING
	orders, err := h.storage.GetOrdersByStatuses([]string{"NEW", "PROCESSING"})
	if err != nil {
		return
	}

	for _, order := range orders {
		accrualResp, err := h.accrualSvc.GetOrderAccrual(order.Number)
		if err != nil {
			if err.Error() == "too many requests" {
				return
			}
			continue
		}

		if accrualResp == nil {
			// Заказ не зарегистрирован, оставляем как есть
			continue
		}

		// Обновить статус
		newStatus := accrualResp.Status
		var accrual *float64
		if accrualResp.Accrual != nil {
			accrual = accrualResp.Accrual
		}

		err = h.storage.UpdateOrderStatus(order.ID, newStatus, accrual)
		if err != nil {
			continue
		}

		// Если PROCESSED и есть accrual, начислить баллы пользователю
		if newStatus == "PROCESSED" && accrual != nil && *accrual > 0 {
			user, err := h.storage.GetUserByID(order.UserID)
			if err != nil {
				continue
			}
			if user != nil {
				newBalance := user.Balance + *accrual
				err = h.storage.UpdateUserBalance(order.UserID, newBalance, user.Withdrawn)
				if err != nil {
					// ignore
				}
			}
		}
	}
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
