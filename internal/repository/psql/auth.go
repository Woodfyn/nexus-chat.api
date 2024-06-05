package psql

import (
	"context"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

type Auth struct {
	db *gorm.DB

	log *logrus.Logger
}

func NewAuth(db *gorm.DB, log *logrus.Logger) *Auth {
	return &Auth{
		db: db,

		log: log,
	}
}

func (a *Auth) CreateUser(ctx context.Context, user *core.User) error {
	return a.db.Create(&user).Error
}

func (a *Auth) GetUserByCredentials(ctx context.Context, phone string) (*core.User, error) {
	var user *core.User
	result := a.db.Where("phone = ?", phone).First(&user)
	if result.RowsAffected == 0 {
		return nil, core.ErrUserNotFound
	} else if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (a *Auth) SetTokenSession(ctx context.Context, input *core.Token) error {
	return a.db.Create(&input).Error
}

func (a *Auth) GetTokenSession(ctx context.Context, refreshToken string) (*core.Token, error) {
	var token *core.Token
	if err := a.db.Where("refresh_token = ?", refreshToken).Find(&token).Error; err != nil {
		return nil, err
	}

	return token, nil
}

func (a *Auth) DeleteTokenSession(ctx context.Context, refreshToken string) error {
	return a.db.Where("refresh_token = ?", refreshToken).Delete(&core.Token{}).Error

}

func (a *Auth) CreateAvatar(ctx context.Context, avatar *core.UserAvatar) error {
	return a.db.Create(&avatar).Error
}
