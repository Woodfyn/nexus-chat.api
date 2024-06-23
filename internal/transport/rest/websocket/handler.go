package websocket

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/gorilla/websocket"
)

var (
	upgrader = &websocket.Upgrader{
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			w.WriteHeader(status)
			w.Write([]byte(reason.Error()))
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Encoder interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type WSHandler struct {
	encoder Encoder

	ConnMap map[int]*WSClient
}

func NewWebSocketHandler(encoder Encoder) *WSHandler {
	return &WSHandler{
		encoder: encoder,

		ConnMap: make(map[int]*WSClient),
	}
}

func (wsh *WSHandler) Stream(w http.ResponseWriter, r *http.Request, userId int) {
	wsc, err := wsh.initWSClient(w, r, userId)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-wsc.exitCh:
			return
		case event := <-wsc.eventCh:
			if event.ReceiveUserID == wsc.userId {
				if event.Message.UserID == wsc.userId {
					event.Message.Username = "You"
				}

				eventRespBytes, err := json.Marshal(core.EventResponse{
					Header:  event.Header,
					Message: event.Message,
				})
				if err != nil {
					wsc.closeConn()
					panic(err)
				}

				encrypt, err := wsh.encoder.Encrypt(eventRespBytes)
				if err != nil {
					wsc.closeConn()
					panic(err)
				}

				wsc.writeMessage(websocket.TextMessage, []byte(base64.StdEncoding.EncodeToString(encrypt)))
			}
		default:
			continue
		}
	}
}

func (wsh *WSHandler) StopStream(userId int) {
	wsc, ok := wsh.ConnMap[userId]
	if !ok {
		panic(core.ErrStreamNotAvailable)
	}

	wsc.exitCh <- struct{}{}

	wsc.closeConn()

	delete(wsh.ConnMap, userId)
}

func (wsh *WSHandler) OnlineStream(userId int) bool {
	_, ok := wsh.ConnMap[userId]

	return ok
}

func (wsh *WSHandler) AddEvent(userId int, event *core.Event) {
	wsc, ok := wsh.ConnMap[userId]
	if !ok {
		panic(core.ErrStreamNotAvailable)
	}

	wsc.eventCh <- event
}
