package models

import "time"

type Withdrawal struct {
	ID          int       `json:"-" db:"id"`
	UserID      int       `json:"-" db:"user_id"`
	Order       string    `json:"order" db:"order_number"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}
