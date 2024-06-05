package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/sirupsen/logrus"
)

const (
	MAX_ROOM_GROUP_SIZE = 10
)

type WSRepositoryPSQL interface {
	GetUserById(ctx context.Context, userId int) (*core.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*core.User, error)
	GetUserOnChat(ctx context.Context, chatId int) ([]*core.ChatUser, error)
	SaveMessage(ctx context.Context, msg *core.ChatMessage) error
	GetWall(ctx context.Context, userId int) ([]*core.Chat, error)
	CreateChat(ctx context.Context, req *core.Chat) error
	GetMessagesByChatId(ctx context.Context, chatId int) ([]*core.ChatMessage, error)
	JoinChat(ctx context.Context, req *core.ChatUser) error
	LeaveChatGroup(ctx context.Context, req *core.ChatUser) error
	DeleteChat(ctx context.Context, userId, chatId int) error
	UpdateChatGroupAdmin(ctx context.Context, req *core.UpdateGroupChatAdminReq) error
	UpdateChatGroupName(ctx context.Context, r *core.UpdateGroupChatNameReq) error
	GetChatById(ctx context.Context, chatId int) (*core.Chat, error)
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

func (ws *WebSocket) CreateChatGroup(ctx context.Context, req *core.CreateChatGroupReq, adminID int) error {
	chat := &core.Chat{
		Name:      req.Name,
		AdminID:   adminID,
		Type:      core.GroupChatType,
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err := ws.psqlRepo.CreateChat(ctx, chat); err != nil {
		return err
	}

	if err := ws.psqlRepo.JoinChat(ctx, &core.ChatUser{
		UserID: adminID,
		ChatID: chat.ID,
	}); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) CreateChatDefault(ctx context.Context, req *core.CreateDefaultChatReq, userID int) error {
	user, err := ws.psqlRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return err
	}

	chat := &core.Chat{
		Type:      core.DefaultChatType,
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err := ws.psqlRepo.CreateChat(ctx, chat); err != nil {
		return err
	}

	if err := ws.psqlRepo.JoinChat(ctx, &core.ChatUser{
		UserID: userID,
		ChatID: chat.ID,
	}); err != nil {
		return err
	}

	if err := ws.psqlRepo.JoinChat(ctx, &core.ChatUser{
		UserID: user.ID,
		ChatID: chat.ID,
	}); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) JoinChatGroup(ctx context.Context, req *core.ChatUser) (*core.ChatMessage, error) {
	chat, err := ws.psqlRepo.GetChatById(ctx, req.ChatID)
	if err != nil {
		return nil, err
	}

	if chat.Type != core.GroupChatType {
		return nil, core.ErrConnotJoinChat
	}

	usersOnChat, err := ws.psqlRepo.GetUserOnChat(ctx, req.ChatID)
	if err != nil {
		return nil, err
	}

	if len(usersOnChat) >= MAX_ROOM_GROUP_SIZE {
		return nil, core.ErrChatGroupFull
	}

	if err := ws.psqlRepo.JoinChat(ctx, req); err != nil {
		if errors.Is(err, core.ErrRecordNotFound) {
			return nil, core.ErrInvalideChatID
		}

		return nil, err
	}

	user, err := ws.psqlRepo.GetUserById(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	msg := &core.ChatMessage{
		Username:  user.Username,
		UserID:    user.ID,
		ChatID:    req.ChatID,
		Text:      fmt.Sprintf("%s joined the chat", user.Username),
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err = ws.psqlRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) GetUserOnChat(ctx context.Context, chatId int) ([]*core.ChatUser, error) {
	return ws.psqlRepo.GetUserOnChat(ctx, chatId)
}

func (ws *WebSocket) SendMessage(ctx context.Context, req *core.SendMessageReq, userId int) (*core.ChatMessage, error) {
	user, err := ws.psqlRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	msg := &core.ChatMessage{
		Username:  user.Username,
		UserID:    user.ID,
		ChatID:    req.ChatID,
		Text:      req.Text,
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err = ws.psqlRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) DeleteChat(ctx context.Context, userId, chatId int) error {
	return ws.psqlRepo.DeleteChat(ctx, userId, chatId)
}

func (ws *WebSocket) LeaveChatGroup(ctx context.Context, req *core.ChatUser) (*core.ChatMessage, error) {
	if err := ws.psqlRepo.LeaveChatGroup(ctx, req); err != nil {
		return nil, err
	}

	user, err := ws.psqlRepo.GetUserById(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	msg := &core.ChatMessage{
		Username:  user.Username,
		UserID:    user.ID,
		ChatID:    req.ChatID,
		Text:      fmt.Sprintf("%s left the chat", user.Username),
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err = ws.psqlRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) UpdateChatGroupAdmin(ctx context.Context, req *core.UpdateGroupChatAdminReq, userId int) (*core.ChatMessage, error) {
	if err := ws.psqlRepo.UpdateChatGroupAdmin(ctx, req); err != nil {
		return nil, err
	}

	user, err := ws.psqlRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	msg := &core.ChatMessage{
		Username:  user.Username,
		UserID:    user.ID,
		ChatID:    req.ChatID,
		Text:      fmt.Sprintf("%s is new chat admin", user.Username),
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err = ws.psqlRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) UpdateChatGroupName(ctx context.Context, req *core.UpdateGroupChatNameReq, userId int) (*core.ChatMessage, error) {
	if err := ws.psqlRepo.UpdateChatGroupName(ctx, req); err != nil {
		return nil, err
	}

	user, err := ws.psqlRepo.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}

	msg := &core.ChatMessage{
		Username:  user.Username,
		UserID:    user.ID,
		ChatID:    req.ChatID,
		Text:      fmt.Sprintf("%s is new chat name", user.Username),
		CreatedAt: time.Now().Format(time.DateTime),
	}

	if err = ws.psqlRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ws *WebSocket) IsAdmin(ctx context.Context, userId, chatId int) (bool, error) {
	chat, err := ws.psqlRepo.GetChatById(ctx, chatId)
	if err != nil {
		return false, err
	}

	return chat.AdminID == userId, nil
}

func (ws *WebSocket) GetMessages(ctx context.Context, chatId int) ([]*core.ChatMessage, error) {
	return ws.psqlRepo.GetMessagesByChatId(ctx, chatId)
}

func (ws *WebSocket) GetWall(ctx context.Context, userId int) ([]*core.WallChatResp, error) {
	wallChats, err := ws.psqlRepo.GetWall(ctx, userId)
	if err != nil {
		return nil, err
	}

	var response []*core.WallChatResp
	for _, wallChat := range wallChats {
		if wallChat.Type == core.DefaultChatType {
			messages, err := ws.psqlRepo.GetMessagesByChatId(ctx, wallChat.ID)
			if err != nil {
				return nil, err
			} else if len(messages) == 0 {
				usersOnChat, err := ws.GetUserOnChat(ctx, wallChat.ID)
				if err != nil {
					return nil, err
				}

				for _, userOnChat := range usersOnChat {
					if userOnChat.UserID != userId {
						user, err := ws.psqlRepo.GetUserById(ctx, userOnChat.UserID)
						if err != nil {
							return nil, err
						}

						response = append(response, &core.WallChatResp{
							ChatID: wallChat.ID,
							Name:   user.Username,
						})
					}
				}

				continue
			}

			usersOnChat, err := ws.GetUserOnChat(ctx, wallChat.ID)
			if err != nil {
				return nil, err
			}

			for _, userOnChat := range usersOnChat {
				if userOnChat.UserID != userId {
					user, err := ws.psqlRepo.GetUserById(ctx, userOnChat.UserID)
					if err != nil {
						return nil, err
					}

					if messages[len(messages)-1].UserID == userOnChat.UserID {
						response = append(response, &core.WallChatResp{
							ChatID:      wallChat.ID,
							Name:        user.Username,
							LastMessage: messages[len(messages)-1].Text,
						})

					}

					response = append(response, &core.WallChatResp{
						ChatID:      wallChat.ID,
						Name:        user.Username,
						LastMessage: fmt.Sprintf("%s: %s", messages[len(messages)-1].Username, messages[len(messages)-1].Text),
					})
				}
			}

			continue
		}

		messages, err := ws.psqlRepo.GetMessagesByChatId(ctx, wallChat.ID)
		if err != nil {
			return nil, err
		} else if len(messages) == 0 {
			response = append(response, &core.WallChatResp{
				ChatID: wallChat.ID,
				Name:   wallChat.Name,
			})

			continue
		}

		if messages[len(messages)-1].UserID == userId {
			response = append(response, &core.WallChatResp{
				ChatID:      wallChat.ID,
				Name:        wallChat.Name,
				LastMessage: messages[len(messages)-1].Text,
			})

			continue
		}

		response = append(response, &core.WallChatResp{
			ChatID:      wallChat.ID,
			Name:        wallChat.Name,
			LastMessage: fmt.Sprintf("%s: %s", messages[len(messages)-1].Username, messages[len(messages)-1].Text),
		})
	}

	return response, nil
}
