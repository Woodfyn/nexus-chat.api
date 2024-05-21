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
	GetProfile(ctx context.Context, userId int) (*core.GetProfileResponse, error)
	UpdateProfile(ctx context.Context, user *core.User) (*core.User, error)
	GetAvatars(ctx context.Context, userId int) ([]*core.GetAllUSerAvatarsResponse, error)
	UploadAvatar(ctx context.Context, file multipart.File, id int) error
	DeleteAvatar(ctx context.Context, userId int, avatarId int) error
}

type WebSocket interface {
	SendMessage(ctx context.Context, r *core.SendMessageReq, userId int) (*core.RoomMessage, error)
	GetMessages(ctx context.Context, req *core.RoomUser) ([]*core.RoomMessage, error)
	GetWall(ctx context.Context, userId int) ([]*core.WallRoomResponse, error)
	CreateRoomGroup(ctx context.Context, r *core.CreateRoomGroupReq, adminID int) error
	CreateRoomUser(ctx context.Context, r *core.CreateRoomUserReq, userID int) error
	JoinRoom(ctx context.Context, r *core.RoomUser) (*core.RoomMessage, error)
	LeaveRoom(ctx context.Context, r *core.RoomUser) (*core.RoomMessage, error)
	DeleteRoom(ctx context.Context, r *core.RoomUser) error
	TransferRoomGroupAdmin(ctx context.Context, r *core.TransferRoomGroupAdminReq, userID int) error
	GetUserOnRoom(ctx context.Context, roomId int) ([]*core.RoomUser, error)
}

type Handler struct {
	authService    Auth
	profileService Profile
	wsService      WebSocket

	eventCh chan []byte

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

		eventCh: make(chan []byte),

		log: deps.Log,
	}
}

func (h *Handler) InitRouter(api *mux.Router) *mux.Router {

	h.initRoomRouter(api)
	h.initProfileRouter(api)
	h.initAuthRouter(api)

	return api
}
