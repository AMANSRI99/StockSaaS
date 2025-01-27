package domain

import (
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	HashedPassword    string    `json:"-"`
	Name              string    `json:"name"`
	ZerodhaAPIKey     string    `json:"zerodha_api_key"`
	ZerodhaAPISecret  string    `json:"zerodha_api_secret"`
	ZerodhaAccessToken string   `json:"zerodha_access_token"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (u *User) SetPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashedBytes)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}

