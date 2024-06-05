package psql

import (
	"context"

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

func (ws *WebSocket) CreateChat(ctx context.Context, req *core.Chat) error {
	return ws.db.Create(&req).Error
}

func (ws *WebSocket) GetUserOnChat(ctx context.Context, chatId int) ([]*core.ChatUser, error) {
	var chatUsers []*core.ChatUser
	if err := ws.db.Model(core.ChatUser{}).Where("chat_id = ?", chatId).Find(&chatUsers).Error; err != nil {
		return nil, err
	}

	return chatUsers, nil
}

func (ws *WebSocket) GetUserByPhone(ctx context.Context, phone string) (*core.User, error) {
	var user *core.User
	if err := ws.db.First(&user, "phone = ?", phone).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (ws *WebSocket) JoinChat(ctx context.Context, req *core.ChatUser) error {
	return ws.db.Where("user_id = ? AND chat_id = ?", req.UserID, req.ChatID).FirstOrCreate(req).Error
}

func (ws *WebSocket) GetWall(ctx context.Context, userId int) ([]*core.Chat, error) {
	var chatUsers []*core.ChatUser
	if err := ws.db.Model(core.ChatUser{}).Where("user_id = ?", userId).Find(&chatUsers).Error; err != nil {
		return nil, err
	}

	var chats []*core.Chat
	for _, chatUser := range chatUsers {
		var searchRoom *core.Chat
		if err := ws.db.First(&searchRoom, chatUser.ChatID).Error; err != nil {
			return nil, err
		}

		chats = append(chats, searchRoom)
	}

	return chats, nil
}

func (ws *WebSocket) GetMessagesByChatId(ctx context.Context, chatId int) ([]*core.ChatMessage, error) {
	var messages []*core.ChatMessage
	if err := ws.db.Model(core.ChatMessage{}).Where("chat_id = ?", chatId).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func (ws *WebSocket) SaveMessage(ctx context.Context, msg *core.ChatMessage) error {
	return ws.db.Create(&msg).Error
}

func (ws *WebSocket) GetUserById(ctx context.Context, userId int) (*core.User, error) {
	var user *core.User
	if err := ws.db.First(&user, "id = ?", userId).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (ws *WebSocket) LeaveChatGroup(ctx context.Context, req *core.ChatUser) error {
	return ws.db.Model(core.ChatUser{}).Where("user_id = ? AND chat_id = ?", req.UserID, req.ChatID).Delete(&core.ChatUser{}).Error
}

func (ws *WebSocket) DeleteChat(ctx context.Context, userId, chatId int) error {
	if err := ws.db.Transaction(func(tx *gorm.DB) error {
		var txChats []*core.ChatUser
		if err := tx.Model(core.ChatUser{}).Where("chat_id = ?", chatId).Find(&txChats).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, chatUser := range txChats {
			if err := tx.Where("user_id = ? AND chat_id = ?", chatUser.UserID, chatUser.ChatID).Delete(&core.ChatUser{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		var messages []*core.ChatMessage
		if err := tx.Model(core.ChatMessage{}).Where("chat_id = ?", chatId).Find(&messages).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, message := range messages {
			if err := tx.Where("id = ?", message.ID).Delete(&core.ChatMessage{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		if err := tx.Where("id = ?", chatId).Delete(&core.Chat{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) UpdateChatGroupAdmin(ctx context.Context, req *core.UpdateGroupChatAdminReq) error {
	return ws.db.Model(core.Chat{}).Where("id = ?", req.ChatID).Update("admin_id", req.NewAdminID).Error
}

func (ws *WebSocket) UpdateChatGroupName(ctx context.Context, req *core.UpdateGroupChatNameReq) error {
	return ws.db.Model(core.Chat{}).Where("id = ?", req.ChatID).Update("name", req.Name).Error
}

func (ws *WebSocket) GetChatById(ctx context.Context, chatId int) (*core.Chat, error) {
	var chat *core.Chat
	if err := ws.db.First(&chat, "id = ?", chatId).Error; err != nil {
		return nil, err
	}

	return chat, nil
}
