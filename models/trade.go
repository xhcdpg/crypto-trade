package models

import (
	"time"
)

type Trade struct {
	ID        string
	Symbol    string
	BuyerID   string
	SellerID  string
	Price     float64
	Quantity  float64
	Timestamp time.Time
}
