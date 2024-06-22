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
			h.newErrorResponse(w, http.StatusUnauthorized, core.ErrEmptyHeader.Error())
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			h.newErrorResponse(w, http.StatusUnauthorized, core.ErrInvalidHeader.Error())
			return
		}

		if len(headerParts[1]) == 0 {
			h.newErrorResponse(w, http.StatusUnauthorized, core.ErrEmptyAccessToken.Error())
			return
		}

		id, err := h.authService.ParseToken(headerParts[1])
		if err != nil {
			h.newErrorResponse(w, http.StatusUnauthorized, core.ErrInvalidAccessToken.Error())
			return
		}

		if !h.authService.IsTokenExpired(headerParts[1]) {
			h.newErrorResponse(w, http.StatusUnauthorized, core.ErrExpiredAccessToken.Error())
			return
		}

		idInt, err := strconv.Atoi(id)
		if err != nil {
			h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "userId", idInt)

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) StreamMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, ok := r.Context().Value("userId").(int)
		if !ok {
			h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
			return
		}

		if ok = h.wsHandler.OnlineStream(userId); !ok {
			h.newErrorResponse(w, http.StatusBadRequest, core.ErrStreamNotAvailable.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "userId", userId)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) LoggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.log.WithFields(logrus.Fields{"method": r.Method, "path": r.URL.Path}).Info()

		handler.ServeHTTP(w, r)
	})
}
