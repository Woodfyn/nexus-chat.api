package rest

import (
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/Woodfyn/chat-api-backend-go/internal/transport/rest/websocket"
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
		NewErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	conn, err := websocket.NewWebSocket(w, r, userId)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	conn.Stream()

	NewResponse(w, http.StatusOK, "stream disconnected")
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
		NewErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	websocket.StopStream(userId)

	NewResponse(w, http.StatusOK, nil)
}
