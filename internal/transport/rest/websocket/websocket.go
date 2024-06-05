package websocket

import (
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

	ConnMap = make(map[int]*WebSocket)
)

type WebSocket struct {
	conn *websocket.Conn

	userId int

	eventCh chan []byte
	exitCh  chan struct{}
}

func NewWebSocket(w http.ResponseWriter, r *http.Request, userId int) (*WebSocket, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	ws, ok := ConnMap[userId]
	if !ok {
		ws := &WebSocket{
			conn: conn,

			userId: userId,

			eventCh: make(chan []byte),
			exitCh:  make(chan struct{}),
		}

		ConnMap[userId] = ws

		return ws, nil
	}

	return ws, nil
}

func (ws *WebSocket) writeMessage(messageType int, data []byte) {
	if err := ws.conn.WriteMessage(messageType, data); err != nil {
		ws.closeConn()
		panic(err)
	}
}

func (ws *WebSocket) closeConn() {
	close(ws.eventCh)
	close(ws.exitCh)
	delete(ConnMap, ws.userId)
	ws.conn.Close()
}

func (ws *WebSocket) Stream() {
	for {
		select {
		case <-ws.exitCh:
			return
		case eventBytes := <-ws.eventCh:
			var event *core.Event
			if err := json.Unmarshal(eventBytes, &event); err != nil {
				ws.closeConn()
				panic(err)
			}

			if event.ReceiveUserID == ws.userId {
				if event.Message.UserID == ws.userId {
					event.Message.Username = "You"
				}

				eventRespBytes, err := json.Marshal(core.EventResponse{
					Header:  event.Header,
					Message: event.Message,
				})
				if err != nil {
					ws.closeConn()
					panic(err)
				}

				ws.writeMessage(websocket.TextMessage, eventRespBytes)
			}
		default:
			continue
		}
	}
}

func StopStream(userId int) {
	ws, ok := ConnMap[userId]
	if !ok {
		return
	}

	ws.exitCh <- struct{}{}

	ws.closeConn()
}
