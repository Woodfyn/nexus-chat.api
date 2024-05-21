package core

type User struct {
	ID          int          `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone       string       `gorm:"unique" json:"phone"`
	Username    string       `gorm:"unique" json:"username"`
	UserAvatars []UserAvatar `json:"image"`
}

type GetProfileResponse struct {
	ID        int    `json:"user_id"`
	Phone     string `json:"phone"`
	Username  string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

type RoomUser struct {
	UserID int
	RoomID int
}

type UserAvatar struct {
	ID     int `gorm:"primaryKey;autoIncrement"`
	UserId int
}

type GetAllUSerAvatarsResponse struct {
	ID        int    `json:"avatar_id"`
	UserId    int    `json:"user_id"`
	AvatarUrl string `json:"avatar_url"`
}

type UpdateUserReq struct {
	Phone    string `json:"phone" validate:"required"`
	Username string `json:"name" validate:"required,gte=2"`
}

func (u *UpdateUserReq) Validate() error {
	return validate.Struct(u)
}
