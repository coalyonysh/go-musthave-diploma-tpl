package storage

import "coalyonysh/go-musthave-diploma-tpl/internal/models"

type Storage interface {
	CreateUser(login, passwordHash string) (*models.User, error)
	GetUserByLogin(login string) (*models.User, error)
	GetUserByID(userID int) (*models.User, error)
	UpdateUserBalance(userID int, balance, withdrawn float64) error

	CreateOrder(userID int, number string) (*models.Order, error)
	GetOrderByNumber(number string) (*models.Order, error)
	GetOrdersByUserID(userID int) ([]models.Order, error)
	GetOrdersByStatuses(statuses []string) ([]models.Order, error)
	UpdateOrderStatus(orderID int, status string, accrual *float64) error

	CreateWithdrawal(userID int, order string, sum float64) error
	GetWithdrawalsByUserID(userID int) ([]models.Withdrawal, error)

	Close() error
}
