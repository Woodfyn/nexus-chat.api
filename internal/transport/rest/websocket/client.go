package websocket

import (
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn *websocket.Conn

	userId int

	eventCh chan *core.Event
	exitCh  chan struct{}
}

func (wsh *WSHandler) initWSClient(w http.ResponseWriter, r *http.Request, userId int) (*WSClient, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	wscFirst, ok := wsh.ConnMap[userId]
	if !ok {
		wscSecond := &WSClient{
			conn: conn,

			userId: userId,

			eventCh: make(chan *core.Event),
			exitCh:  make(chan struct{}),
		}

		wsh.ConnMap[userId] = wscSecond

		return wscSecond, nil
	}

	return wscFirst, nil
}

func (wsc *WSClient) closeConn() {
	close(wsc.eventCh)
	close(wsc.exitCh)
	wsc.conn.Close()
}

func (wsc *WSClient) writeMessage(messageType int, data []byte) {
	if err := wsc.conn.WriteMessage(messageType, data); err != nil {
		wsc.closeConn()
		panic(err)
	}
}
