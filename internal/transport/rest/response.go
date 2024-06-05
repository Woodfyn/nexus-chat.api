package rest

import (
	"encoding/json"
	"net/http"
)

func NewResponse(w http.ResponseWriter, status int, input any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if input == nil {
		return nil
	}

	respose, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = w.Write(respose)
	if err != nil {
		return err
	}

	return nil
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(w http.ResponseWriter, status int, msg string) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	msgErr := errorResponse{
		Error: msg,
	}

	respose, err := json.Marshal(msgErr)
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
