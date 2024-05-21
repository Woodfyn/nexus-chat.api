package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/Woodfyn/chat-api-backend-go/pkg/token"
	"github.com/Woodfyn/chat-api-backend-go/pkg/verife"
	"github.com/sirupsen/logrus"
)

type AuthRepositoryPSQL interface {
	CreateUser(ctx context.Context, user *core.User) (*core.User, error)
	GetUserByCredentials(ctx context.Context, phone string) (*core.User, error)
	SetTokenSession(ctx context.Context, input *core.Token) error
	GetTokenSession(ctx context.Context, refreshToken string) (*core.Token, error)
	DeleteTokenSession(ctx context.Context, refreshToken string) error
	CreateAvatar(ctx context.Context, avatar *core.UserAvatar) (*core.UserAvatar, error)
}

type VerifyRepositoryREDIS interface {
	SetCode(ctx context.Context, id int, code string) error
	Verify(ctx context.Context, code string) (string, error)
}

type VerifyRepositoryTWILIO interface {
	SendCode(ctx context.Context, code string, phone string) error
}

type Auth struct {
	psqlRepo   AuthRepositoryPSQL
	redisRepo  VerifyRepositoryREDIS
	twilioRepo VerifyRepositoryTWILIO

	manager         *token.Manager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	log *logrus.Logger
}

func NewAuth(psqlRepo AuthRepositoryPSQL, redisRepo VerifyRepositoryREDIS, twilioRepo VerifyRepositoryTWILIO, manager *token.Manager, accessTokenTTL time.Duration, refreshTokenTTL time.Duration, log *logrus.Logger) *Auth {
	return &Auth{
		psqlRepo:   psqlRepo,
		redisRepo:  redisRepo,
		twilioRepo: twilioRepo,

		manager:         manager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,

		log: log,
	}
}

func (a *Auth) Register(ctx context.Context, user *core.AuthRegister) error {
	_, err := a.psqlRepo.CreateUser(ctx, &core.User{
		Phone:    user.Phone,
		Username: user.Username,
	})
	if err != nil {
		return core.ErrThisCredIsAlready
	}

	return nil
}

func (a *Auth) Login(ctx context.Context, auth *core.AuthLogin) error {
	user, err := a.psqlRepo.GetUserByCredentials(ctx, auth.Phone)
	if err != nil {
		return err
	}

	code := verife.GenereteCode()

	if err := a.redisRepo.SetCode(ctx, user.ID, code); err != nil {
		return err
	}

	if err := a.twilioRepo.SendCode(ctx, code, auth.Phone); err != nil {
		return err
	}

	return nil
}

func (a *Auth) Verify(ctx context.Context, code string) ([]string, error) {
	id, err := a.redisRepo.Verify(ctx, code)
	if err != nil {
		return nil, err
	}

	tokens, err := a.genereteTokens(id)
	if err != nil {
		return nil, err
	}

	userIdInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	err = a.psqlRepo.SetTokenSession(ctx, &core.Token{
		UserID:       userIdInt,
		RefreshToken: tokens[0],
		CreatedAt:    time.Now().Format(time.DateTime),
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string) ([]string, error) {
	tokenSession, err := a.psqlRepo.GetTokenSession(ctx, refreshToken)
	if errors.Is(err, core.ErrRefreshTokenNotFound) {
		return nil, core.ErrRefreshTokenIsExpired
	} else if err != nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.DateTime, tokenSession.CreatedAt)
	if err != nil {
		return nil, err
	}

	expirationTime := createdAt.Add(a.refreshTokenTTL)

	if time.Now().After(expirationTime) {
		return nil, core.ErrRefreshTokenIsExpired
	}

	if err := a.psqlRepo.DeleteTokenSession(ctx, refreshToken); err != nil {
		return nil, err
	}

	tokens, err := a.genereteTokens(strconv.Itoa(tokenSession.UserID))
	if err != nil {
		return nil, err
	}

	err = a.psqlRepo.SetTokenSession(ctx, &core.Token{
		UserID:       tokenSession.UserID,
		RefreshToken: tokens[0],
		CreatedAt:    time.Now().Format(time.DateTime),
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (a *Auth) genereteTokens(userId string) ([]string, error) {
	accessToken, err := a.manager.NewJWT(userId, a.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := a.manager.NewRefreshToken()
	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0, 2)
	tokens = append(tokens, refreshToken)
	tokens = append(tokens, accessToken)

	return tokens, nil
}

func (a *Auth) ParseToken(token string) (string, error) {
	id, err := a.manager.Parse(token)

	return id, err
}

func (a *Auth) IsTokenExpired(token string) bool {
	return a.manager.IsTokenExpired(token)
}
