package core

var (
	DefaultChatType = "default-chat"
	GroupChatType   = "group-chat"
)

type Chat struct {
	ID        int `gorm:"primaryKey;autoIncrement"`
	Name      string
	AdminID   int
	Type      string
	CreatedAt string
	Users     []ChatUser    `gorm:"constraint:OnDelete:CASCADE;"`
	Messages  []ChatMessage `gorm:"constraint:OnDelete:CASCADE;"`
}

type CreateChatGroupReq struct {
	Name string `json:"chat_name" validate:"required"`
}

type CreateDefaultChatReq struct {
	Phone string `json:"phone" validate:"required"`
}

type UpdateGroupChatAdminReq struct {
	NewAdminID int `json:"new_admin_id" validate:"required"`
	ChatID     int `json:"chat_id" validate:"required"`
}

type UpdateGroupChatNameReq struct {
	Name   string `json:"new_chat_name" validate:"required"`
	ChatID int    `json:"chat_id" validate:"required"`
}

type JoinChatGroupReq struct {
	ChatID int `json:"chat_id" validate:"required"`
}

type WallChatsResp struct {
	Data []WallChatResp `json:"data"`
}

type WallChatResp struct {
	ChatID      int    `json:"chat_id"`
	Name        string `json:"chat_name"`
	LastMessage string `json:"chat_last_message"`
}
