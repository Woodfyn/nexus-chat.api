package rest

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"

	"github.com/sirupsen/logrus"
)

func (h *Handler) AuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			NewResponse(w, http.StatusUnauthorized, core.ErrEmptyHeader.Error())
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			NewResponse(w, http.StatusUnauthorized, core.ErrInvalidHeader.Error())
			return
		}

		if len(headerParts[1]) == 0 {
			NewResponse(w, http.StatusUnauthorized, core.ErrEmptyAccessToken.Error())
			return
		}

		id, err := h.authService.ParseToken(headerParts[1])
		if err != nil {
			NewResponse(w, http.StatusUnauthorized, core.ErrInvalidAccessToken.Error())
			return
		}

		if !h.authService.IsTokenExpired(headerParts[1]) {
			NewResponse(w, http.StatusUnauthorized, core.ErrExpiredAccessToken.Error())
			return
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			NewResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "userId", idInt)

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) LoggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.log.WithFields(logrus.Fields{"method": r.Method, "path": r.URL.Path}).Info()

		handler.ServeHTTP(w, r)
	})
}
