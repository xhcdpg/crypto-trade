package models

import (
	"github.com/xhcdpg/crypto-trade/types"
	"time"
)

type Order struct {
	ID         string
	UserID     string
	Symbol     string
	Side       types.Side
	Type       types.OrderType
	Leverage   uint
	Quantity   float64
	Price      float64 // 委托价
	StopPrice  float64 // 触发价/止盈价/止损价
	Status     types.OrderStatus
	MarginType types.MarginMode
	Timestamp  time.Time
}
