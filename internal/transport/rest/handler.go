package rest

import (
	"context"
	"mime/multipart"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Auth interface {
	Register(ctx context.Context, user *core.AuthRegister) error
	Login(ctx context.Context, auth *core.AuthLogin) error
	Refresh(ctx context.Context, refreshToken string) ([]string, error)
	Verify(ctx context.Context, code string) ([]string, error)
	ParseToken(token string) (string, error)
	IsTokenExpired(accessToken string) bool
}

type Profile interface {
	GetProfile(ctx context.Context, userId int) (*core.GetProfileResp, error)
	UpdateProfile(ctx context.Context, user *core.User) error
	GetAvatars(ctx context.Context, userId int) ([]*core.GetAllUserAvatarsResp, error)
	UploadAvatar(ctx context.Context, file multipart.File, userId int) error
	DeleteAvatar(ctx context.Context, userId int, avatarId int) error
}

type WebSocket interface {
	CreateChatGroup(ctx context.Context, req *core.CreateChatGroupReq, adminID int) error
	LeaveChatGroup(ctx context.Context, req *core.ChatUser) (*core.ChatMessage, error)
	JoinChatGroup(ctx context.Context, req *core.ChatUser) (*core.ChatMessage, error)
	UpdateChatGroupName(ctx context.Context, req *core.UpdateGroupChatNameReq, userId int) (*core.ChatMessage, error)
	UpdateChatGroupAdmin(ctx context.Context, req *core.UpdateGroupChatAdminReq, userId int) (*core.ChatMessage, error)
	CreateChatDefault(ctx context.Context, req *core.CreateDefaultChatReq, userID int) error
	DeleteChat(ctx context.Context, userId, chatId int) error
	GetUserOnChat(ctx context.Context, chatId int) ([]*core.ChatUser, error)
	GetWall(ctx context.Context, userId int) ([]*core.WallChatResp, error)
	SendMessage(ctx context.Context, req *core.SendMessageReq, userId int) (*core.ChatMessage, error)
	GetMessages(ctx context.Context, chatId int) ([]*core.ChatMessage, error)
	IsAdmin(ctx context.Context, userId int, chatId int) (bool, error)
}

type Handler struct {
	authService    Auth
	profileService Profile
	wsService      WebSocket

	log *logrus.Logger
}

type Deps struct {
	Auth      Auth
	Profile   Profile
	WebSocket WebSocket

	Log *logrus.Logger
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		authService:    deps.Auth,
		profileService: deps.Profile,
		wsService:      deps.WebSocket,

		log: deps.Log,
	}
}

func (h *Handler) InitRouter(api *mux.Router) *mux.Router {
	h.initAuthRouter(api)
	h.initProfileRouter(api)
	h.initStreamRouter(api)
	h.initWebSocketRouter(api)

	return api
}
