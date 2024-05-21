package core

var (
	NewMessageEventHeader = "NewMessage"
	JoinRoomEventHeader   = "JoinRoom"
	LeaveRoomEventHeader  = "LeaveRoom"
)

type Event struct {
	Header        string
	Message       RoomMessage
	ReceiveUserID int
}

type EventResponse struct {
	Header  string
	Message RoomMessage
}
