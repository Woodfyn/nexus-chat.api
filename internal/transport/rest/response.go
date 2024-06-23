package rest

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type tokenResponse struct {
	Token string `json:"token"`
}

func (h *Handler) newResponse(w http.ResponseWriter, status int, input any) error {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status)

	if input == nil {
		return nil
	}

	respose, err := json.Marshal(input)
	if err != nil {
		return err
	}

	encryptResp, err := h.encoder.Encrypt(respose)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(base64.StdEncoding.EncodeToString(encryptResp)))
	if err != nil {
		return err
	}

	return nil
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) newErrorResponse(w http.ResponseWriter, status int, msg string) error {
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
