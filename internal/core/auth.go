package core

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type AuthRegister struct {
	Phone    string `json:"phone" validate:"required"`
	Username string `json:"username" validate:"required,gte=2"`
}

func (a *AuthRegister) Validate() error {
	return validate.Struct(a)
}

type AuthLogin struct {
	Phone string `json:"phone" validate:"required"`
}

func (a *AuthLogin) Validate() error {
	return validate.Struct(a)
}

type Token struct {
	ID           int `gorm:"primaryKey;autoIncrement"`
	UserID       int
	RefreshToken string
	CreatedAt    string
}
