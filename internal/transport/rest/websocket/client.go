package websocket

import (
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn

	userId int

	eventCh chan *core.Event
	exitCh  chan struct{}
}

func (h *Handler) initClient(w http.ResponseWriter, r *http.Request, userId int) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	wscFirst, ok := h.ConnMap[userId]
	if !ok {
		wscSecond := &Client{
			conn: conn,

			userId: userId,

			eventCh: make(chan *core.Event),
			exitCh:  make(chan struct{}),
		}

		h.ConnMap[userId] = wscSecond

		return wscSecond, nil
	}

	return wscFirst, nil
}

func (c *Client) closeConn() {
	close(c.eventCh)
	close(c.exitCh)
	c.conn.Close()
}

func (c *Client) writeMessage(messageType int, data []byte) {
	if err := c.conn.WriteMessage(messageType, data); err != nil {
		c.closeConn()
		panic(err)
	}
}
