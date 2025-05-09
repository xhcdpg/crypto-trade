package models

import "time"

type FoundingPayment struct {
	UserID string
	Symbol string
	Amount float64
	Time   time.Time
}
