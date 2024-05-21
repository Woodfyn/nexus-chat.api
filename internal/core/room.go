package core

var (
	UserRoomType  = "user_room"
	GroupRoomType = "group_room"
)

type Room struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"unique"`
	AdminID   int
	Type      string
	CreatedAt string
	Users     []RoomUser    `gorm:"constraint:OnDelete:CASCADE;"`
	Messages  []RoomMessage `gorm:"constraint:OnDelete:CASCADE;"`
}

type CreateRoomGroupReq struct {
	Name string `json:"name" validate:"required"`
}

func (c *CreateRoomGroupReq) Validate() error {
	return validate.Struct(c)
}

type CreateRoomUserReq struct {
	Phone string `json:"name" validate:"required"`
}

func (c *CreateRoomUserReq) Validate() error {
	return validate.Struct(c)
}

type TransferRoomGroupAdminReq struct {
	UserId int `json:"user_id" validate:"required"`
	RoomId int `json:"room_id" validate:"required"`
}

func (c *TransferRoomGroupAdminReq) Validate() error {
	return validate.Struct(c)
}

type JoinRoomReq struct {
	UserId int
	RoomId int
}

type RoomWallResponse struct {
	Data []WallRoomResponse `json:"data"`
}

type WallRoomResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	LastMessage string `json:"last_message"`
}
