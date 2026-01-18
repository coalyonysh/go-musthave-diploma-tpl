package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"coalyonysh/go-musthave-diploma-tpl/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(databaseURI string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(context.Background(), databaseURI)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) CreateUser(login, passwordHash string) (*models.User, error) {
	query := `
		INSERT INTO users (login, password_hash, balance, withdrawn, created_at)
		VALUES ($1, $2, 0, 0, NOW())
		RETURNING id, login, balance, withdrawn, created_at`

	user := &models.User{}
	err := s.pool.QueryRow(context.Background(), query, login, passwordHash).Scan(
		&user.ID, &user.Login, &user.Balance, &user.Withdrawn, &user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) GetUserByLogin(login string) (*models.User, error) {
	query := `SELECT id, login, password_hash, balance, withdrawn, created_at FROM users WHERE login = $1`

	user := &models.User{}
	err := s.pool.QueryRow(context.Background(), query, login).Scan(
		&user.ID, &user.Login, &user.Password, &user.Balance, &user.Withdrawn, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) GetUserByID(userID int) (*models.User, error) {
	query := `SELECT id, login, password_hash, balance, withdrawn, created_at FROM users WHERE id = $1`

	user := &models.User{}
	err := s.pool.QueryRow(context.Background(), query, userID).Scan(
		&user.ID, &user.Login, &user.Password, &user.Balance, &user.Withdrawn, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) UpdateUserBalance(userID int, balance, withdrawn float64) error {
	query := `UPDATE users SET balance = $1, withdrawn = $2 WHERE id = $3`
	_, err := s.pool.Exec(context.Background(), query, balance, withdrawn, userID)
	return err
}

func (s *PostgresStorage) CreateOrder(userID int, number string) (*models.Order, error) {
	query := `
		INSERT INTO orders (user_id, number, status, uploaded_at)
		VALUES ($1, $2, 'NEW', NOW())
		RETURNING id, user_id, number, status, accrual, uploaded_at`

	order := &models.Order{}
	err := s.pool.QueryRow(context.Background(), query, userID, number).Scan(
		&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *PostgresStorage) GetOrderByNumber(number string) (*models.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`

	order := &models.Order{}
	err := s.pool.QueryRow(context.Background(), query, number).Scan(
		&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return order, nil
}

func (s *PostgresStorage) GetOrdersByUserID(userID int) ([]models.Order, error) {
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStorage) GetOrdersByStatuses(statuses []string) ([]models.Order, error) {
	if len(statuses) == 0 {
		return []models.Order{}, nil
	}

	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE status = ANY($1)`

	rows, err := s.pool.Query(context.Background(), query, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStorage) UpdateOrderStatus(orderID int, status string, accrual *float64) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`
	_, err := s.pool.Exec(context.Background(), query, status, accrual, orderID)
	return err
}

func (s *PostgresStorage) CreateWithdrawal(userID int, order string, sum float64) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, NOW())`
	_, err := s.pool.Exec(context.Background(), query, userID, order, sum)
	return err
}

func (s *PostgresStorage) GetWithdrawalsByUserID(userID int) ([]models.Withdrawal, error) {
	query := `SELECT id, user_id, order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

	rows, err := s.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		err := rows.Scan(&w.ID, &w.UserID, &w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	return withdrawals, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

// InitDB создает таблицы, если они не существуют
func (s *PostgresStorage) InitDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			balance DECIMAL(10,2) DEFAULT 0,
			withdrawn DECIMAL(10,2) DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id),
			number VARCHAR(255) UNIQUE NOT NULL,
			status VARCHAR(50) DEFAULT 'NEW',
			accrual DECIMAL(10,2),
			uploaded_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id),
			order_number VARCHAR(255) NOT NULL,
			sum DECIMAL(10,2) NOT NULL,
			processed_at TIMESTAMP DEFAULT NOW()
		)`,
	}

	for _, query := range queries {
		_, err := s.pool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}
