package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"

	"github.com/gorilla/mux"
)

func (h *Handler) initWebSocketRouter(api *mux.Router) {
	chat := api.PathPrefix("/chat").Subrouter()
	{
		chat.Use(h.AuthMiddleware, h.StreamMiddleware)

		groupChat := chat.PathPrefix("/group").Subrouter()
		{
			groupChat.HandleFunc("/create", h.wsCreateChatGroup).Methods(http.MethodPost)
			groupChat.HandleFunc("/join", h.wsJoinChatGroup).Methods(http.MethodPost)
			groupChat.HandleFunc("/leave/{chatId}", h.wsLeaveChatGroup).Methods(http.MethodDelete)

			admin := groupChat.PathPrefix("/admin").Subrouter()
			{
				admin.HandleFunc("/update", h.wsUpdateChatGroupAdmin).Methods(http.MethodPut)
				admin.HandleFunc("/update/name", h.wsUpdateChatGroupName).Methods(http.MethodPut)
				admin.HandleFunc("/delete/{chatId}", h.wsDeleteChatGroup).Methods(http.MethodDelete)
			}
		}

		defaultChat := chat.PathPrefix("/default").Subrouter()
		{
			defaultChat.HandleFunc("/create", h.wsCreateChatDefault).Methods(http.MethodPost)
			defaultChat.HandleFunc("/delete/{chatId}", h.wsDeleteChatDefault).Methods(http.MethodDelete)
		}

		chat.HandleFunc("/wall", h.wsWall).Methods(http.MethodGet)

		msg := chat.PathPrefix("/message").Subrouter()
		{
			msg.HandleFunc("/send", h.wsSendMessage).Methods(http.MethodPost)
			msg.HandleFunc("/get/{chatId}", h.wsGetMessages).Methods(http.MethodGet)
		}
	}

}

// @Summary CreateChatGroup
// @Tags Chat
// @Security ApiKeyAuth
// @Description create chat group
// @ID createChatGroup
// @Accept json
// @Produce json
// @Param input body core.CreateChatGroupReq true "create chat group"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/group/create [post]
func (h *Handler) wsCreateChatGroup(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	var req *core.CreateChatGroupReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err = h.wsService.CreateChatGroup(r.Context(), req, userId); err != nil {
		if errors.Is(err, core.ErrDuplicatedKey) {
			h.newErrorResponse(w, http.StatusBadRequest, core.ErrThisCredIsAlready.Error())
			return
		}

		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary CreateChatDefault
// @Tags Chat
// @Security ApiKeyAuth
// @Description create chat default
// @ID createChatDefault
// @Accept json
// @Produce json
// @Param input body core.CreateDefaultChatReq true "create chat default"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/default/create [post]
func (h *Handler) wsCreateChatDefault(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	var req *core.CreateDefaultChatReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err = h.wsService.CreateChatDefault(r.Context(), req, userId); err != nil {
		if errors.Is(err, core.ErrDuplicatedKey) {
			h.newErrorResponse(w, http.StatusBadRequest, core.ErrThisCredIsAlready.Error())
			return
		}

		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary JoinChat
// @Tags Chat
// @Security ApiKeyAuth
// @Description join chat
// @ID joinChat
// @Produce json
// @Param chatId path string true "chat id"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/group/join/{chatId} [post]
func (h *Handler) wsJoinChatGroup(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	var req *core.JoinChatGroupReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	msg, err := h.wsService.JoinChatGroup(r.Context(), &core.ChatUser{
		UserID: userId,
		ChatID: req.ChatID,
	})

	switch err {
	case nil:
		chatUsers, err := h.wsService.GetUserOnChat(r.Context(), req.ChatID)
		if err != nil {
			h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, chatUser := range chatUsers {
			event := &core.Event{
				Header:        core.JoinChatEventHeader,
				Message:       msg,
				ReceiveUserID: chatUser.UserID,
			}

			h.wsHandler.AddEvent(chatUser.UserID, event)
		}

		h.newResponse(w, http.StatusOK, nil)
		return
	case core.ErrChatGroupFull, core.ErrConnotJoinChat, core.ErrInvalideChatID:
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	default:
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// @Summary LeaveChatGroup
// @Tags Chat
// @Security ApiKeyAuth
// @Description leave chat group
// @ID leaveChatGroup
// @Produce json
// @Param chatId path string true "chat id"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/group/leave/{chatId} [delete]
func (h *Handler) wsLeaveChatGroup(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	chatId, err := getChatIdFromRequest(r)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err = h.wsService.IsAdmin(r.Context(), userId, chatId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	} else if ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrAdminCannnotLeave.Error())
		return
	}

	msg, err := h.wsService.LeaveChatGroup(r.Context(), &core.ChatUser{
		UserID: userId,
		ChatID: chatId,
	})
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	chatUsers, err := h.wsService.GetUserOnChat(r.Context(), chatId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, chatUser := range chatUsers {
		event := &core.Event{
			Header:        core.LeaveChatGroupEventHeader,
			Message:       msg,
			ReceiveUserID: chatUser.UserID,
		}

		h.wsHandler.AddEvent(chatUser.UserID, event)
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary DeleteChatGroup
// @Tags Chat
// @Security ApiKeyAuth
// @Description delete chat group
// @ID deleteChatGroup
// @Produce json
// @Param chatId path string true "chat id"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/group/admin/delete/{chatId} [delete]
func (h *Handler) wsDeleteChatGroup(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	chatId, err := getChatIdFromRequest(r)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ok, err = h.wsService.IsAdmin(r.Context(), userId, chatId)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrNotAdmin.Error())
		return
	} else if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.wsService.DeleteChat(r.Context(), userId, chatId); err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary DeleteChatDefault
// @Tags Chat
// @Security ApiKeyAuth
// @Description delete chat default
// @ID deleteChatDefault
// @Produce json
// @Param chatId path string true "chat id"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/default/delete/{chatId} [delete]
func (h *Handler) wsDeleteChatDefault(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	chatId, err := getChatIdFromRequest(r)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.wsService.DeleteChat(r.Context(), userId, chatId); err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary UpdateChatGroupName
// @Tags Chat
// @Security ApiKeyAuth
// @Description update chat group name
// @ID updateChatGroupName
// @Accept json
// @Produce json
// @Param input body core.UpdateGroupChatNameReq true "update chat group name"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/group/admin/update [put]
func (h *Handler) wsUpdateChatGroupName(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	var req *core.UpdateGroupChatNameReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	ok, err = h.wsService.IsAdmin(r.Context(), userId, req.ChatID)
	if !ok {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	} else if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrNotAdmin.Error())
		return
	}

	msg, err := h.wsService.UpdateChatGroupName(r.Context(), req, userId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	cahtUsers, err := h.wsService.GetUserOnChat(r.Context(), req.ChatID)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, chatUser := range cahtUsers {
		event := &core.Event{
			Header:        core.UpdateChatGroupName,
			Message:       msg,
			ReceiveUserID: chatUser.UserID,
		}

		h.wsHandler.AddEvent(chatUser.UserID, event)
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary TransferChatGroupAdmin
// @Tags Chat
// @Security ApiKeyAuth
// @Description transfer chat group admin
// @ID transferChatGroupAdmin
// @Accept json
// @Produce json
// @Param input body core.UpdateGroupChatAdminReq true "update chat group admin"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/group/admin/update [put]
func (h *Handler) wsUpdateChatGroupAdmin(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	var req *core.UpdateGroupChatAdminReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	ok, err = h.wsService.IsAdmin(r.Context(), userId, req.ChatID)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	} else if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrNotAdmin.Error())
		return
	}

	msg, err := h.wsService.UpdateChatGroupAdmin(r.Context(), req, userId)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cahtUsers, err := h.wsService.GetUserOnChat(r.Context(), req.ChatID)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, chatUser := range cahtUsers {
		event := &core.Event{
			Header:        core.UpdateChatGroupAdmin,
			Message:       msg,
			ReceiveUserID: chatUser.UserID,
		}

		h.wsHandler.AddEvent(chatUser.UserID, event)
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary ChatWall
// @Tags Chat
// @Security ApiKeyAuth
// @Description get chat wall
// @ID chatWall
// @Produce json
// @Success 200 {array} core.WallChatResp
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/wall [get]
func (h *Handler) wsWall(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	response, err := h.wsService.GetWall(r.Context(), userId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, response)
}

// @Summary SendMessage
// @Tags Chat
// @Security ApiKeyAuth
// @Description send message
// @ID sendMessage
// @Produce json
// @Accept json
// @Param message body core.SendMessageReq true "message"
// @Seccess 200
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/message/send [post]
func (h *Handler) wsSendMessage(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	var req *core.SendMessageReq
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	chatUsers, err := h.wsService.GetUserOnChat(r.Context(), req.ChatID)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(chatUsers) == 0 {
		h.newErrorResponse(w, http.StatusForbidden, core.ErrNoChats.Error())
		return
	}

	for _, chatUser := range chatUsers {
		if req.ChatID != chatUser.ChatID {
			h.newErrorResponse(w, http.StatusBadRequest, core.ErrInvalideChatID.Error())
			return
		}
	}

	msg, err := h.wsService.SendMessage(r.Context(), req, userId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, chatUser := range chatUsers {
		event := &core.Event{
			Header:        core.NewMessageEventHeader,
			Message:       msg,
			ReceiveUserID: chatUser.UserID,
		}

		h.wsHandler.AddEvent(chatUser.UserID, event)
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary GetMessages
// @Tags Chat
// @Security ApiKeyAuth
// @Description get messages from chat
// @ID getMessages
// @Produce json
// @Param roomId path string true "chat id"
// @Seccess 200 {array} core.ChatMessage
// @Failure 400,500 {object} errorResponse
// @Router /api/chat/message/get/{chatId} [get]
func (h *Handler) wsGetMessages(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	chatId, err := getChatIdFromRequest(r)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := h.wsService.GetMessages(r.Context(), chatId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, message := range messages {
		if message.ID == userId {
			message.Username = "You"
		}
	}

	h.newResponse(w, http.StatusOK, messages)
}

func getChatIdFromRequest(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	chatId := vars["chatId"]
	if chatId == "" {
		return 0, core.ErrEmptyUserID
	}

	chatIdInt, err := strconv.Atoi(chatId)
	if err != nil {
		return 0, err
	}

	return chatIdInt, nil
}
