package core

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Token{}, &Chat{}, &ChatUser{}, &ChatMessage{}, &UserAvatar{})
}
