package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/Woodfyn/chat-api-backend-go/internal/transport/rest/ws"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func (h *Handler) initRoomRouter(api *mux.Router) {
	ws := api.PathPrefix("/ws").Subrouter()
	{
		ws.Use(h.AuthMiddleware)

		ws.HandleFunc("/wall", h.wsWall).Methods(http.MethodGet)
		ws.HandleFunc("/send", h.wsSendMessage).Methods(http.MethodPost)
		ws.HandleFunc("/get/messages/{roomId}", h.wsGetMessages).Methods(http.MethodGet)
		ws.HandleFunc("/join/{roomId}", h.wsJoinRoom).Methods(http.MethodGet)
		ws.HandleFunc("/leave/{roomId}", h.wsLeaveRoom).Methods(http.MethodGet)

		create := ws.PathPrefix("/create").Subrouter()
		{
			create.HandleFunc("/create/group", h.wsCreateRoomGroup).Methods(http.MethodPost)
			create.HandleFunc("/create/user", h.wsCreateRoomUser).Methods(http.MethodPost)
		}

		admin := ws.PathPrefix("/admin").Subrouter()
		{
			admin.Use(nil)

			admin.HandleFunc("/delete/{roomId}", h.wsDeleteRoom).Methods(http.MethodDelete)
			admin.HandleFunc("/update/group-name", h.wsUpdateRoomGroupName).Methods(http.MethodPut)
			admin.HandleFunc("/transfer-admin", h.wsTransferRoomGroupAdmin).Methods(http.MethodPut)
		}

		stream := ws.PathPrefix("/stream").Subrouter()
		{
			stream.HandleFunc("/connect", h.wsStream).Methods(http.MethodGet)
			stream.HandleFunc("/disconnect", h.wsStreamStop).Methods(http.MethodGet)
		}
	}
}

// @Summary Stream
// @Tags Websocket
// @Security ApiKeyAuth
// @Description stream session
// @ID stream
// @Failure 400,500 {object} response
// @Router /ws/stream/ [get]
func (h *Handler) wsStream(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	ws, err := ws.InitWebSocket(w, r)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> InitWebSocket"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	for {
		select {
		case eventBytes := <-h.eventCh:
			var event *core.Event
			if err := json.Unmarshal(eventBytes, &event); err != nil {
				h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Unmarshal"}).Error(err)
				NewResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			if event.ReceiveUserID == userId {
				if event.Message.UserID == userId {
					event.Message.Username = "You"
				}

				eventRespBytes, err := NewEventResponse(event)
				if err != nil {
					h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> NewEventResponse"}).Error(err)
					NewResponse(w, http.StatusInternalServerError, err.Error())
					return
				}

				if err := ws.WriteMessage(websocket.TextMessage, eventRespBytes); err != nil {
					h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> WriteMessage"}).Error(err)
					NewResponse(w, http.StatusInternalServerError, err.Error())
					return
				}
			}
		default:
			continue
		}
	}
}

func (h *Handler) wsStreamStop(w http.ResponseWriter, r *http.Request) {}

// @Summary SendMessage
// @Tags Websocket
// @Security ApiKeyAuth
// @Description send message
// @ID sendMessage
// @Produce json
// @Param message body core.SendMessageReq true "message"
// @Param roomId path string true "room id"
// @Seccess 200
// @Failure 400,500 {object} response
// @Router /ws/send [post]
func (h *Handler) wsSendMessage(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	var req *core.SendMessageReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> ReadAll"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> Unmarshal"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> Validate"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	roomUsers, err := h.wsService.GetUserOnRoom(r.Context(), req.RoomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> GetUserOnRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(roomUsers) == 0 {
		h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> GetUserOnRoom"}).Error(err)
		NewResponse(w, int(http.StatusForbidden), "You have no chats")
		return
	}

	for _, roomUser := range roomUsers {
		if req.RoomId != roomUser.RoomID {
			h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> GetUserOnRoom"}).Error(err)
			NewResponse(w, int(http.StatusForbidden), "You have no chats")
			return
		}
	}

	msg, err := h.wsService.SendMessage(r.Context(), req, userId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> SaveMessage"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, user := range roomUsers {
		event := core.Event{
			Header:        core.NewMessageEventHeader,
			Message:       core.PtrMsgToNonePtrMsg(msg),
			ReceiveUserID: user.UserID,
		}

		eventByte, err := json.Marshal(event)
		if err != nil {
			h.log.WithFields(logrus.Fields{"handler": "wsSendMessage -> Marshal"}).Error(err)
			continue
		}

		h.eventCh <- eventByte
	}

	NewResponse(w, http.StatusOK, "OK")
}

// @Summary GetMessages
// @Tags Websocket
// @Security ApiKeyAuth
// @Description get messages for room
// @ID getMessages
// @Produce json
// @Param roomId path string true "room id"
// @Seccess 200 {array} core.RoomMessage
// @Failure 400,500 {object} response
// @Router /get/messages/{roomId} [get]
func (h *Handler) wsGetMessages(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	roomId, err := getRoomIdFromRequest(r)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req := &core.RoomUser{
		UserID: userId,
		RoomID: roomId,
	}

	messages, err := h.wsService.GetMessages(r.Context(), req)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "wsGetMessages -> GetMessages"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, message := range messages {
		if message.ID == userId {
			message.Username = "You"
		}
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(messages)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileWallRooms -> Marshal"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err := w.Write(response); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileWallRooms -> Write"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// @Summary CreateRoom
// @Tags Websocket
// @Security ApiKeyAuth
// @Description create room
// @ID createRoom
// @Accept json
// @Produce json
// @Param input body core.CreateRoomReq true "create room"
// @Success 200
// @Failure 400,500 {object} response
// @Router /create [post]
func (h *Handler) wsCreateRoomGroup(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	var req *core.CreateRoomGroupReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> ReadAll"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> Unmarshal"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> Validate"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.wsService.CreateRoomGroup(r.Context(), req, userId); err != nil {
		if errors.Is(err, core.ErrDuplicatedKey) {
			h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> CreateRoom"}).Error(err)
			NewResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> CreateRoom"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, "OK")
}

func (h *Handler) wsCreateRoomUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	var req *core.CreateRoomUserReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> ReadAll"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> Unmarshal"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> Validate"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.wsService.CreateRoomUser(r.Context(), req, userId); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> CreateRoom"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, "OK")
}

func (h *Handler) wsTransferRoomGroupAdmin(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	var req *core.TransferRoomGroupAdminReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> ReadAll"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> Unmarshal"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> Validate"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = h.wsService.TransferRoomGroupAdmin(r.Context(), req, userId); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "CreateRoom -> CreateRoom"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, "OK")
}

// @Summary JoinRoom
// @Tags Websocket
// @Security ApiKeyAuth
// @Description join room
// @ID joinRoom
// @Produce json
// @Param roomId path string true "room id"
// @Success 200
// @Failure 400,500 {object} response
// @Router /join/{roomId} [get]
func (h *Handler) wsJoinRoom(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	roomId, err := getRoomIdFromRequest(r)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req := &core.RoomUser{
		UserID: userId,
		RoomID: roomId,
	}

	msg, err := h.wsService.JoinRoom(r.Context(), req)
	if err != nil {
		if errors.Is(err, core.ErrJoinAlreadyExist) {
			h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> JoinRoom"}).Error(err)
			NewResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> JoinRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	roomUsers, err := h.wsService.GetUserOnRoom(r.Context(), roomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> GetUserOnRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, user := range roomUsers {
		event := core.Event{
			Header:        core.JoinRoomEventHeader,
			Message:       core.PtrMsgToNonePtrMsg(msg),
			ReceiveUserID: user.UserID,
		}

		eventByte, err := json.Marshal(event)
		if err != nil {
			h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> Marshal"}).Error(err)
			NewResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		h.eventCh <- eventByte
	}

	NewResponse(w, http.StatusOK, "OK")
}

// @Summary LeaveRoom
// @Tags Websocket
// @Security ApiKeyAuth
// @Description leave room
// @ID leaveRoom
// @Produce json
// @Param roomId path string true "room id"
// @Success 200
// @Failure 400,500 {object} response
// @Router /leave/{roomId} [get]
func (h *Handler) wsLeaveRoom(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	roomId, err := getRoomIdFromRequest(r)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req := &core.RoomUser{
		UserID: userId,
		RoomID: roomId,
	}

	msg, err := h.wsService.LeaveRoom(r.Context(), req)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> JoinRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	roomUsers, err := h.wsService.GetUserOnRoom(r.Context(), roomId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> GetUserOnRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, user := range roomUsers {
		event := core.Event{
			Header:        core.LeaveRoomEventHeader,
			Message:       core.PtrMsgToNonePtrMsg(msg),
			ReceiveUserID: user.UserID,
		}

		eventByte, err := json.Marshal(event)
		if err != nil {
			h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> Marshal"}).Error(err)
			NewResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		h.eventCh <- eventByte
	}

	NewResponse(w, http.StatusOK, "OK")
}

// @Summary DeleteRoom
// @Tags Websocket
// @Security ApiKeyAuth
// @Description delete room
// @ID deleteRoom
// @Produce json
// @Param roomId path string true "room id"
// @Success 200
// @Failure 400,500 {object} response
// @Router /delete/{roomId} [get]
func (h *Handler) wsDeleteRoom(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	roomId, err := getRoomIdFromRequest(r)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> getRoomIdFromRequest"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req := &core.RoomUser{
		UserID: userId,
		RoomID: roomId,
	}

	if err := h.wsService.DeleteRoom(r.Context(), req); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "roomJoinRoom -> JoinRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, "OK")
}

func getRoomIdFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	roomId := vars["roomId"]
	if roomId == "" {
		return 0, core.ErrEmptyRoomId
	}

	roomIdInt, err := strconv.Atoi(roomId)
	if err != nil {
		return 0, err
	}

	return roomIdInt, nil
}

func (h *Handler) wsUpdateRoomGroupName(w http.ResponseWriter, r *http.Request) {}

// @Summary RoomWall
// @Tags Websocket
// @Security ApiKeyAuth
// @Description get room wall
// @ID roomWall
// @Produce json
// @Success 200 {array} core.WallRoomResponse
// @Failure 400,500 {object} response
// @Router /wall [get]
func (h *Handler) wsWall(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileWallRooms -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	response, err := h.wsService.GetWall(r.Context(), userId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileWallRooms -> GetRoomsWall"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileWallRooms -> Marshal"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err := w.Write(responseBytes); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileWallRooms -> Write"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
