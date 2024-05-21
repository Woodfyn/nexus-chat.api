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
	ErrEmptyUserId       = errors.New("user id is empty")
	ErrEmptyUsername     = errors.New("username is empty")
	ErrThisCredIsAlready = errors.New("this cred is already")
	ErrRecordNotFound    = gorm.ErrRecordNotFound

	ErrCodeMismatch = errors.New("code mismatch")
	ErrCodeNotFound = errors.New("code not found")

	ErrEmptyCode = errors.New("code is empty")

	ErrDuplicatedKey = gorm.ErrDuplicatedKey

	ErrEmptyRoomId = errors.New("room id is empty")

	ErrNoneMessage = errors.New("none message")

	ErrJoinAlreadyExist = errors.New("user already in room")

	ErrEmptyAvatarId  = errors.New("avatar id is empty")
	ErrAvatarNotFound = errors.New("avatar not found")
)
