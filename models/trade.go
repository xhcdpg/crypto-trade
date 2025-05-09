package models

import (
	"github.com/xhcdpg/crypto-trade/types"
	"time"
)

type Trade struct {
	ID        string
	Symbol    string
	Side      types.Side
	BuyerID   string
	SellerID  string
	Price     float64
	Quantity  float64
	Timestamp time.Time
}
