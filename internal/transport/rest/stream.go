package rest

import (
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/gorilla/mux"
)

func (h *Handler) initStreamRouter(api *mux.Router) {
	stream := api.PathPrefix("/stream").Subrouter()
	{
		stream.Use(h.AuthMiddleware)

		stream.HandleFunc("/connect", h.wsStreamConnect).Methods(http.MethodGet)
		stream.HandleFunc("/disconnect", h.wsStreamDisconnect).Methods(http.MethodPost)
	}
}

// @Summary Connect
// @Tags Stream
// @Security ApiKeyAuth
// @Description connect to a streaming session
// @ID streamConnect
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/stream/connect [get]
func (h *Handler) wsStreamConnect(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	h.wsHandler.Stream(w, r, userId)

	h.newResponse(w, http.StatusOK, "stream disconnected")
}

// @Summary Disconnect
// @Tags Stream
// @Security ApiKeyAuth
// @Description disconnect from the streaming session
// @ID streamDisconnect
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/stream/disconnect [post]
func (h *Handler) wsStreamDisconnect(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	h.wsHandler.StopStream(userId)

	h.newResponse(w, http.StatusOK, nil)
}
