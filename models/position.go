package models

import "github.com/xhcdpg/crypto-trade/types"

type Position struct {
	ID                string
	UserID            string
	Symbol            string
	Side              types.Side
	ContractType      string
	Leverage          uint
	EntryPrice        float64
	Quantity          float64
	AllocatedMargin   float64
	MarkPrice         float64
	UnrealizedPnl     float64
	RealizedPnl       float64
	MaintenanceMargin float64
	InitialMargin     float64
	LiquidationPrice  float64
}
