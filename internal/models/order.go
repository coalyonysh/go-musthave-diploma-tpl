package models

import "time"

type Order struct {
	ID         int       `json:"-" db:"id"`
	UserID     int       `json:"-" db:"user_id"`
	Number     string    `json:"number" db:"number"`
	Status     string    `json:"status" db:"status"`
	Accrual    *float64  `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
