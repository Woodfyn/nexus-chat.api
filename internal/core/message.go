package core

type RoomMessage struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string `json:"username"`
	UserID    int    `json:"user_id"`
	RoomID    int    `json:"room_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type SendMessageReq struct {
	RoomId int    `json:"room_id" validate:"required"`
	Text   string `json:"text" validate:"required"`
}

func (s *SendMessageReq) Validate() error {
	return validate.Struct(s)
}

func PtrMsgToNonePtrMsg(event *RoomMessage) RoomMessage {
	return RoomMessage{
		ID:        event.ID,
		Username:  event.Username,
		UserID:    event.UserID,
		RoomID:    event.RoomID,
		Text:      event.Text,
		CreatedAt: event.CreatedAt,
	}
}
