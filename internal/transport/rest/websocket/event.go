package websocket

import (
	"encoding/json"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
)

func AddEvent(userId int, event core.Event) error {
	ws, ok := ConnMap[userId]
	if !ok {
		return core.ErrStreamNotAvailable
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ws.eventCh <- eventBytes

	return nil
}
