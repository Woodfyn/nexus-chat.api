package psql

import (
	"context"
	"errors"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

type WebSocket struct {
	db *gorm.DB

	log *logrus.Logger
}

func NewWebSocket(db *gorm.DB, log *logrus.Logger) *WebSocket {
	return &WebSocket{
		db: db,

		log: log,
	}
}

func (ws *WebSocket) CreateRoomGroup(ctx context.Context, r *core.Room) error {
	if err := ws.db.Create(r).Error; err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) CreateRoomUser(ctx context.Context, r *core.Room) (int, error) {
	if err := ws.db.Create(r).Error; err != nil {
		return 0, err
	}

	return r.ID, nil
}

func (ws *WebSocket) GetUserOnRoom(ctx context.Context, roomId int) ([]*core.RoomUser, error) {
	var roomUsers []*core.RoomUser

	if err := ws.db.Model(core.RoomUser{}).Where("room_id = ?", roomId).Find(&roomUsers).Error; err != nil {
		return nil, err
	}

	return roomUsers, nil
}

func (ws *WebSocket) GetUserByPhone(ctx context.Context, phone string) (*core.User, error) {
	var user *core.User
	if err := ws.db.First(&user, "phone = ?", phone).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (ws *WebSocket) JoinRoom(ctx context.Context, r *core.RoomUser) error {
	if err := ws.db.Where("room_id = ? AND user_id = ?", r.RoomID, r.UserID).First(&core.RoomUser{}).Error; err == nil {
		return core.ErrJoinAlreadyExist
	}

	var roomModel core.Room
	if err := ws.db.First(&roomModel, "id = ?", r.RoomID).Error; err != nil {
		return err
	}

	var user core.User
	if err := ws.db.First(&user, "id = ?", r.UserID).Error; err != nil {
		return err
	}

	var roomUser core.RoomUser
	if err := ws.db.First(&roomUser, "user_id = ? AND room_id = ?", user.ID, roomModel.ID).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	roomUser = core.RoomUser{
		UserID: user.ID,
		RoomID: roomModel.ID,
	}

	if err := ws.db.Create(&roomUser).Error; err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) GetWall(ctx context.Context, userId int) ([]*core.Room, error) {
	roomUsers := []*core.RoomUser{}
	if err := ws.db.Model(core.RoomUser{}).Where("user_id = ?", userId).Find(&roomUsers).Error; err != nil {
		return nil, err
	}

	rooms := []*core.Room{}
	for _, roomUser := range roomUsers {
		var room *core.Room
		if err := ws.db.First(&room, roomUser.RoomID).Error; err != nil {
			return nil, err
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (ws *WebSocket) SaveMessage(ctx context.Context, msg *core.RoomMessage) (*core.RoomMessage, error) {
	if err := ws.db.Create(&msg).Error; err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) GetUserById(ctx context.Context, userId int) (*core.User, error) {
	var user *core.User

	if err := ws.db.First(&user, "id = ?", userId).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (ws *WebSocket) GetMessagesByRoomId(ctx context.Context, roomId int) ([]*core.RoomMessage, error) {
	var messages []*core.RoomMessage

	if err := ws.db.Model(core.RoomMessage{}).Where("room_id = ?", roomId).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func (ws *WebSocket) LeaveRoom(ctx context.Context, r *core.RoomUser) error {
	if err := ws.db.Model(core.RoomUser{}).Where("user_id = ? AND room_id = ?", r.UserID, r.RoomID).Delete(&core.RoomUser{}).Error; err != nil {
		return err
	}

	return nil
}
func (ws *WebSocket) DeleteRoom(ctx context.Context, r *core.RoomUser) error {
	if err := ws.db.Transaction(func(tx *gorm.DB) error {
		var room []*core.RoomUser
		if err := tx.Model(core.RoomUser{}).Where("room_id = ?", r.RoomID).Find(&room).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, roomUser := range room {
			if err := tx.Where("user_id = ? AND room_id = ?", roomUser.UserID, roomUser.RoomID).Delete(&core.RoomUser{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		var messages []*core.RoomMessage
		if err := tx.Model(core.RoomMessage{}).Where("room_id = ?", r.RoomID).Find(&messages).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, message := range messages {
			if err := tx.Where("id = ?", message.ID).Delete(&core.RoomMessage{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		if err := tx.Where("id = ?", r.RoomID).Delete(&core.Room{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
