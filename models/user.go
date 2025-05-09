package models

import "github.com/xhcdpg/crypto-trade/types"

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	TotalBalance float64
	MarginMode   types.MarginType
	Positions    []Position
}
