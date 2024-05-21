package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/sirupsen/logrus"
)

const (
	MAX_ROOM_USER_SIZE  = 2
	MAX_ROOM_GROUP_SIZE = 10
)

type WSRepositoryPSQL interface {
	GetUserById(ctx context.Context, userId int) (*core.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*core.User, error)
	GetUserOnRoom(ctx context.Context, roomId int) ([]*core.RoomUser, error)
	SaveMessage(ctx context.Context, msg *core.RoomMessage) (*core.RoomMessage, error)
	GetWall(ctx context.Context, userId int) ([]*core.Room, error)
	CreateRoomGroup(ctx context.Context, r *core.Room) error
	CreateRoomUser(ctx context.Context, r *core.Room) (int, error)
	GetMessagesByRoomId(ctx context.Context, roomId int) ([]*core.RoomMessage, error)
	JoinRoom(ctx context.Context, r *core.RoomUser) error
	LeaveRoom(ctx context.Context, r *core.RoomUser) error
	DeleteRoom(ctx context.Context, r *core.RoomUser) error
}

type WebSocket struct {
	psqlRepo WSRepositoryPSQL

	log *logrus.Logger
}

func NewWebSocket(psqlRepo WSRepositoryPSQL, log *logrus.Logger) *WebSocket {
	return &WebSocket{
		psqlRepo: psqlRepo,

		log: log,
	}
}

func (ws *WebSocket) SendMessage(ctx context.Context, r *core.SendMessageReq, userId int) (*core.RoomMessage, error) {
	user, err := ws.psqlRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	msgWithId, err := ws.psqlRepo.SaveMessage(ctx, &core.RoomMessage{
		Username:  user.Username,
		UserID:    user.ID,
		RoomID:    r.RoomId,
		Text:      r.Text,
		CreatedAt: time.Now().Format(time.DateTime),
	})
	if err != nil {
		return nil, err
	}

	return msgWithId, nil
}

func (ws *WebSocket) GetUserOnRoom(ctx context.Context, roomId int) ([]*core.RoomUser, error) {
	return ws.psqlRepo.GetUserOnRoom(ctx, roomId)
}

func (ws *WebSocket) GetMessages(ctx context.Context, req *core.RoomUser) ([]*core.RoomMessage, error) {
	return ws.psqlRepo.GetMessagesByRoomId(ctx, req.RoomID)
}

func (ws *WebSocket) CreateRoomGroup(ctx context.Context, r *core.CreateRoomGroupReq, adminID int) error {
	if err := ws.psqlRepo.CreateRoomGroup(ctx, &core.Room{
		Name:      r.Name,
		AdminID:   adminID,
		Type:      core.GroupRoomType,
		CreatedAt: time.Now().Format(time.DateTime),
	}); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) CreateRoomUser(ctx context.Context, r *core.CreateRoomUserReq, userId int) error {
	user, err := ws.psqlRepo.GetUserByPhone(ctx, r.Phone)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%d_%d", user.ID, userId)

	roomId, err := ws.psqlRepo.CreateRoomUser(ctx, &core.Room{
		Name:      name,
		AdminID:   userId,
		Type:      core.UserRoomType,
		CreatedAt: time.Now().Format(time.DateTime),
	})
	if err != nil {
		return err
	}

	if err := ws.psqlRepo.JoinRoom(ctx, &core.RoomUser{
		UserID: user.ID,
		RoomID: roomId,
	}); err != nil {
		return err
	}

	if err := ws.psqlRepo.JoinRoom(ctx, &core.RoomUser{
		UserID: userId,
		RoomID: roomId,
	}); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) JoinRoom(ctx context.Context, r *core.RoomUser) (*core.RoomMessage, error) {
	if err := ws.psqlRepo.JoinRoom(ctx, r); err != nil {
		return nil, err
	}

	user, err := ws.psqlRepo.GetUserById(ctx, r.UserID)
	if err != nil {
		return nil, err
	}

	msg, err := ws.psqlRepo.SaveMessage(ctx, &core.RoomMessage{
		Username:  user.Username,
		UserID:    user.ID,
		RoomID:    r.RoomID,
		Text:      fmt.Sprintf("%s joined the chat", user.Username),
		CreatedAt: time.Now().Format(time.DateTime),
	})
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) DeleteRoom(ctx context.Context, r *core.RoomUser) error {
	return ws.psqlRepo.DeleteRoom(ctx, r)
}

func (ws *WebSocket) LeaveRoom(ctx context.Context, r *core.RoomUser) (*core.RoomMessage, error) {
	if err := ws.psqlRepo.LeaveRoom(ctx, r); err != nil {
		return nil, err
	}

	user, err := ws.psqlRepo.GetUserById(ctx, r.UserID)
	if err != nil {
		return nil, err
	}

	msg, err := ws.psqlRepo.SaveMessage(ctx, &core.RoomMessage{
		Username:  user.Username,
		UserID:    user.ID,
		RoomID:    r.RoomID,
		Text:      fmt.Sprintf("%s left the chat", user.Username),
		CreatedAt: time.Now().Format(time.DateTime),
	})
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) TransferRoomGroupAdmin(ctx context.Context, r *core.TransferRoomGroupAdminReq, userID int) error {

	return nil
}

func (ws *WebSocket) GetWall(ctx context.Context, userId int) ([]*core.WallRoomResponse, error) {
	rooms, err := ws.psqlRepo.GetWall(ctx, userId)
	if err != nil {
		return nil, err
	}

	var response []*core.WallRoomResponse
	for _, room := range rooms {
		messages, err := ws.psqlRepo.GetMessagesByRoomId(ctx, room.ID)
		if err != nil {
			return nil, err
		}

		lastMsg := searchLastMsg(messages)

		response = append(response, &core.WallRoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			LastMessage: lastMsg.Text,
		})
	}

	return response, nil
}

func searchLastMsg(messages []*core.RoomMessage) *core.RoomMessage {
	var lastMsg *core.RoomMessage
	var maxId = 0

	for _, msg := range messages {
		if msg.ID > maxId {
			maxId = msg.ID
			lastMsg = msg
		}
	}

	return lastMsg
}
