package models

import "time"

type User struct {
	ID        int       `json:"-" db:"id"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"password,omitempty" db:"password_hash"`
	Balance   float64   `json:"current" db:"balance"`
	Withdrawn float64   `json:"withdrawn" db:"withdrawn"`
	CreatedAt time.Time `json:"-" db:"created_at"`
}
