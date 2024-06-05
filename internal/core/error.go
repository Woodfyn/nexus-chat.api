package core

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrEmptyHeader   = errors.New("empty auth header")
	ErrInvalidHeader = errors.New("invalid auth header")

	ErrEmptyAccessToken   = errors.New("access token is empty")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrExpiredAccessToken = errors.New("expired access token")

	ErrEmptyRefreshToken     = errors.New("refresh token is empty")
	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenIsExpired = errors.New("refresh token not found")

	ErrUserNotFound      = errors.New("user not found")
	ErrEmptyUserID       = errors.New("user id is empty")
	ErrEmptyUsername     = errors.New("username is empty")
	ErrThisCredIsAlready = errors.New("this cred is already")
	ErrRecordNotFound    = gorm.ErrRecordNotFound

	ErrCodeNotFound = errors.New("code not found")

	ErrEmptyCode = errors.New("code is empty")

	ErrDuplicatedKey = gorm.ErrDuplicatedKey

	ErrEmptyChatID    = errors.New("chat id is empty")
	ErrInvalideChatID = errors.New("invalid chat id")

	ErrNoneMessage = errors.New("none message")

	ErrEmptyAvatarID  = errors.New("avatar id is empty")
	ErrAvatarNotFound = errors.New("avatar not found")

	ErrStreamNotAvailable = errors.New("stream not available")
	ErrStreamIsClosed     = errors.New("stream is closed")

	ErrNotAdmin          = errors.New("you are not admin")
	ErrAdminCannnotLeave = errors.New("admin cannot leave")

	ErrNoChats        = errors.New("you have no chats")
	ErrConnotJoinChat = errors.New("cannot join chat")
	ErrJoinIsAlready  = errors.New("join is already")
	ErrChatGroupFull  = errors.New("chat group is full")
)
