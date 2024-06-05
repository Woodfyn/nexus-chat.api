package core

type User struct {
	ID          int          `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Phone       string       `gorm:"unique" json:"phone"`
	Username    string       `gorm:"unique" json:"username"`
	UserAvatars []UserAvatar `json:"image"`
}
type ChatUser struct {
	UserID int `gorm:"primaryKey"`
	ChatID int `gorm:"primaryKey"`
}

type UserAvatar struct {
	ID     int `gorm:"primaryKey;autoIncrement"`
	UserID int
}

type GetProfileResp struct {
	ID        int    `json:"user_id"`
	Phone     string `json:"phone"`
	Username  string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

type GetAllUserAvatarsResp struct {
	ID        int    `json:"avatar_id"`
	AvatarUrl string `json:"avatar_url"`
}

type UpdateUserReq struct {
	Phone    string `json:"phone" validate:"required"`
	Username string `json:"name" validate:"required,gte=2"`
}
