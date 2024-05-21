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
	var user core.User
	if err := p.db.First(&user, "id = ?", userId).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *Profile) UpdateProfile(ctx context.Context, user *core.User) (*core.User, error) {
	var existingUser core.User
	if err := p.db.First(&existingUser, "id = ?", user.ID).Error; err != nil {
		return nil, err
	}

	updatedUser := user

	if err := p.db.Model(&updatedUser).Updates(&updatedUser).Error; err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (p *Profile) UploadAvatar(ctx context.Context, avatar *core.UserAvatar) (*core.UserAvatar, error) {
	if err := p.db.Create(&avatar).Error; err != nil {
		return nil, err
	}

	return avatar, nil
}

func (p *Profile) GetAvatars(ctx context.Context, userId int) ([]*core.UserAvatar, error) {
	var avatars []*core.UserAvatar

	if err := p.db.Where("user_id = ?", userId).Find(&avatars).Error; err != nil {
		return nil, err
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

	if err := p.db.Delete(&core.UserAvatar{}, id).Error; err != nil {
		return err
	}

	return nil
}
