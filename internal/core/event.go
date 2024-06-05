package core

var (
	NewMessageEventHeader     = "NewMessage"
	JoinChatEventHeader       = "JoinChat"
	LeaveChatGroupEventHeader = "LeaveChatGroup"
	UpdateChatGroupAdmin      = "UpdateChatGroupAdmin"
	UpdateChatGroupName       = "UpdateChatGroupName"
)

type Event struct {
	Header        string
	Message       *ChatMessage
	ReceiveUserID int
}

type EventResponse struct {
	Header  string
	Message *ChatMessage
}
