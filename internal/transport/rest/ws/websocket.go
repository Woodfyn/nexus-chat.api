package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type webSocket struct {
	conn *websocket.Conn
}

var upgrader = &websocket.Upgrader{
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

func InitWebSocket(w http.ResponseWriter, r *http.Request) (*webSocket, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &webSocket{
		conn: conn,
	}, nil
}

func (w *webSocket) WriteMessage(messageType int, data []byte) error {
	if err := w.conn.WriteMessage(messageType, data); err != nil {
		w.conn.Close()

		return err
	}

	return nil
}

func (w *webSocket) Close() {
	w.conn.Close()
}
