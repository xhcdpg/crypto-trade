package user

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/xhcdpg/crypto-trade/models"
	"github.com/xhcdpg/crypto-trade/types"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *sql.DB
}

var GlobalUserService *UserService

func NewUserService(db *sql.DB) *UserService {
	service := &UserService{
		db: db,
	}
	GlobalUserService = service
	return service
}

func (u *UserService) RegisterUser(username, email, password string, marginTMode types.MarginMode) error {
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userID := uuid.New().String()
	_, err = u.db.Exec("INSERT INTO users(id,username,email,password_hashed,total_balance,margin_mode) VALUES($1,$2,$3,$4,$5,$6)", userID, username, email, string(passwordHashed), 0.0, marginTMode)

	return err
}

func (u *UserService) GetUser(userID string) (*models.User, error) {
	var (
		user        models.User
		marginTMode string
	)

	err := u.db.QueryRow("SELECT id,username,email,password_hashed,total_balance,margin_mode FROM users WHERE id=$1", userID).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.TotalBalance, &marginTMode)
	if err != nil {
		return nil, err
	}
	user.MarginMode = types.MarginMode(marginTMode)

	// todo
	user.Positions = []models.Position{}

	return &user, nil
}

func (u *UserService) Deposit(user *models.User, amount float64) {
	user.TotalBalance += amount
	u.db.Exec("UPDATE users SET total_balance = $1 WHERE id = $2", user.TotalBalance, user.ID)
}

func (u *UserService) AddMarginToPosition(user *models.User, positionID string, amount float64) error {
	if user.TotalBalance < amount {
		return errors.New("insufficient balance")
	}
	if user.MarginMode == types.CrossMargin {
		u.Deposit(user, amount)
		return nil
	}
	for i, pos := range user.Positions {
		if pos.ID == positionID {
			user.Positions[i].AllocatedMargin += amount
			user.TotalBalance -= amount
			u.db.Exec("UPDATE users SET total_balance = $1 WHERE id = $2", user.TotalBalance, user.ID)
			return nil
		}
	}

	return errors.New("position not found")
}
