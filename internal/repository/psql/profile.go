package psql

import (
	"context"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

type Profile struct {
	db *gorm.DB

	log *logrus.Logger
}

func NewProfile(db *gorm.DB, log *logrus.Logger) *Profile {
	return &Profile{
		db: db,

		log: log,
	}
}

func (p *Profile) GetProfile(ctx context.Context, userId int) (*core.User, error) {
	var user *core.User
	result := p.db.Where("id = ?", userId).First(&user)
	if result.RowsAffected == 0 {
		return nil, core.ErrUserNotFound
	} else if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (p *Profile) UpdateProfile(ctx context.Context, user *core.User) error {
	return p.db.Model(&core.User{}).Where("id = ?", user.ID).Updates(&user).Error
}

func (p *Profile) SaveAvatar(ctx context.Context, avatar *core.UserAvatar) error {
	return p.db.Create(&avatar).Error
}

func (p *Profile) GetAvatars(ctx context.Context, userId int) ([]*core.UserAvatar, error) {
	var avatars []*core.UserAvatar
	if result := p.db.Where("user_id = ?", userId).Find(&avatars); result.RowsAffected == 0 {
		return nil, core.ErrAvatarNotFound
	} else if result.Error != nil {
		return nil, result.Error
	}

	return avatars, nil
}

func (p *Profile) GetAvatar(ctx context.Context, avatarId int) (*core.UserAvatar, error) {
	var avatar core.UserAvatar
	if err := p.db.Where("id = ?", avatarId).First(&avatar).Error; err != nil {
		return nil, err
	}

	return &avatar, nil
}

func (p *Profile) DeleteAvatar(ctx context.Context, id int) error {
	if err := p.db.Where("id = ?", id).Delete(&core.UserAvatar{}).Error; err != nil {
		return err
	}

	return nil
}
