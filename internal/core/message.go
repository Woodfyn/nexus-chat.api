package core

type ChatMessage struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"chat_message_id"`
	Username  string `json:"username"`
	UserID    int    `json:"user_id"`
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type SendMessageReq struct {
	ChatID int    `json:"chat_id" validate:"required"`
	Text   string `json:"text" validate:"required"`
}

func PtrMsgToNonePtrMsg(event *ChatMessage) ChatMessage {
	return ChatMessage{
		ID:        event.ID,
		Username:  event.Username,
		UserID:    event.UserID,
		ChatID:    event.ChatID,
		Text:      event.Text,
		CreatedAt: event.CreatedAt,
	}
}
