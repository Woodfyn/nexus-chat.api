package core

type Token struct {
	ID           int `gorm:"primaryKey;autoIncrement"`
	UserID       int
	RefreshToken string
	CreatedAt    string
}

type AuthRegister struct {
	Phone    string `json:"phone" validate:"required"`
	Username string `json:"username" validate:"required,gte=2"`
}

type AuthLogin struct {
	Phone string `json:"phone" validate:"required"`
}
