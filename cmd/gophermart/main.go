package main

import (
	"log"
	"time"

	"coalyonysh/go-musthave-diploma-tpl/internal/config"
	"coalyonysh/go-musthave-diploma-tpl/internal/handlers"
	"coalyonysh/go-musthave-diploma-tpl/internal/middleware"
	"coalyonysh/go-musthave-diploma-tpl/internal/services"
	"coalyonysh/go-musthave-diploma-tpl/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Инициализация хранилища
	storage, err := storage.NewPostgresStorage(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer storage.Close()

	// Инициализация таблиц
	if err := storage.InitDB(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Accrual service
	accrualSvc := services.NewAccrualService(cfg.AccrualSystemAddress)

	// Handlers
	authHandler := handlers.NewAuthHandler(storage)
	ordersHandler := handlers.NewOrdersHandler(storage)
	balanceHandler := handlers.NewBalanceHandler(storage)

	// Router
	r := gin.Default()

	// Routes
	api := r.Group("/api/user")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)

		// Protected routes
		protected := api.Group("", middleware.AuthMiddleware())
		{
			protected.POST("/orders", ordersHandler.UploadOrder)
			protected.GET("/orders", ordersHandler.GetOrders)
			protected.GET("/balance", balanceHandler.GetBalance)
			protected.POST("/balance/withdraw", balanceHandler.Withdraw)
			protected.GET("/withdrawals", balanceHandler.GetWithdrawals)
		}
	}

	// Фоновый процесс для обновления статусов заказов
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			updateOrderStatuses(storage, accrualSvc)
		}
	}()

	log.Printf("Server starting on %s", cfg.RunAddress)
	if err := r.Run(cfg.RunAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func updateOrderStatuses(storage storage.Storage, accrualSvc *services.AccrualService) {
	// Получить заказы в NEW и PROCESSING
	orders, err := storage.GetOrdersByStatuses([]string{"NEW", "PROCESSING"})
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
		return
	}

	for _, order := range orders {
		accrualResp, err := accrualSvc.GetOrderAccrual(order.Number)
		if err != nil {
			if err.Error() == "too many requests" {
				log.Println("Too many requests to accrual service, skipping...")
				return
			}
			log.Printf("Failed to get accrual for order %s: %v", order.Number, err)
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

		err = storage.UpdateOrderStatus(order.ID, newStatus, accrual)
		if err != nil {
			log.Printf("Failed to update order %s: %v", order.Number, err)
			continue
		}

		// Если PROCESSED и есть accrual, начислить баллы пользователю
		if newStatus == "PROCESSED" && accrual != nil && *accrual > 0 {
			user, err := storage.GetUserByID(order.UserID)
			if err != nil {
				log.Printf("Failed to get user %d: %v", order.UserID, err)
				continue
			}
			if user != nil {
				newBalance := user.Balance + *accrual
				err = storage.UpdateUserBalance(order.UserID, newBalance, user.Withdrawn)
				if err != nil {
					log.Printf("Failed to update balance for user %d: %v", order.UserID, err)
				}
			}
		}
	}
}
