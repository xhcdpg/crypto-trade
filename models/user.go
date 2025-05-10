package models

import "github.com/xhcdpg/crypto-trade/types"

type User struct {
	ID           string           `json:"id"`
	Username     string           `json:"username"`
	Email        string           `json:"email"`
	PasswordHash string           `json:"password_hash"`
	TotalBalance float64          `json:"total_balance"`
	MarginMode   types.MarginMode `json:"margin_mode"`
	Positions    []Position       `json:"positions"`
}
