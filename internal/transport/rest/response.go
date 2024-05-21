package rest

import (
	"encoding/json"
	"net/http"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"
)

type response struct {
	Message string `json:"message"`
}

func NewResponse(w http.ResponseWriter, status int, message string) error {
	if message == "OK" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status)

		return nil
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	msg := response{Message: message}

	respose, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = w.Write(respose)
	if err != nil {
		return err
	}

	return nil
}

type tokenResponse struct {
	Token string `json:"token"`
}

func NewTokenResponse(w http.ResponseWriter, status int, token string) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	msg := tokenResponse{Token: token}

	respose, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = w.Write(respose)
	if err != nil {
		return err
	}

	return nil
}

func NewEventResponse(event *core.Event) ([]byte, error) {
	response := &core.EventResponse{
		Header:  event.Header,
		Message: event.Message,
	}

	return json.Marshal(response)
}
