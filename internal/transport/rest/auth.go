package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"

	"github.com/gorilla/mux"
)

func (h *Handler) initAuthRouter(r *mux.Router) {
	api := r.PathPrefix("/auth").Subrouter()
	{
		api.HandleFunc("/register", h.authRegister).Methods(http.MethodPost)
		api.HandleFunc("/login", h.authLogin).Methods(http.MethodPost)
		api.HandleFunc("/verify/{code}", h.authVerify).Methods(http.MethodGet)
		api.HandleFunc("/refresh", h.authRefresh).Methods(http.MethodPost)
	}
}

// @Summary Register
// @Tags Auth
// @Description register
// @ID register
// @Accept json
// @Produce json
// @Param input body core.AuthRegister true "credentials"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/auth/register [post]
func (h *Handler) authRegister(w http.ResponseWriter, r *http.Request) {
	var auth *core.AuthRegister
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &auth); err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := auth.Validate(); err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err := h.authService.Register(r.Context(), auth); err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, nil)
}

// @Summary Login
// @Tags Auth
// @Description login
// @ID login
// @Accept json
// @Produce json
// @Param input body core.AuthLogin true "credentials"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/auth/login [post]
func (h *Handler) authLogin(w http.ResponseWriter, r *http.Request) {
	var auth *core.AuthLogin
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := json.Unmarshal(reqBytes, &auth); err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := auth.Validate(); err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err = h.authService.Login(r.Context(), auth); err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			NewErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		NewErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, nil)
}

// @Summary Verify
// @Tags Auth
// @Description verify
// @ID verify
// @Accept  json
// @Produce  json
// @Param code path string true "code"
// @Success 200 {object} tokenResponse
// @Failure 400,500 {object} errorResponse
// @Router /api/auth/verify/{code} [get]
func (h *Handler) authVerify(w http.ResponseWriter, r *http.Request) {
	code, err := getCodeFromRequest(r)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Verify(r.Context(), code)
	if err != nil {
		if errors.Is(err, core.ErrCodeNotFound) {
			NewErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		NewErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + tokens[0],
		HttpOnly: true,
		MaxAge:   3600,
		Path:     "/",
	})

	NewTokenResponse(w, http.StatusOK, tokens[1])
}

// @Summary Refresh
// @Tags Auth
// @Description refresh
// @ID refresh
// @Accept json
// @Produce json
// @Success 200 {object} tokenResponse
// @Failure 400,500 {object} errorResponse
// @Router /api/auth/refresh [post]
func (h *Handler) authRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	refreshToken, err := getTokenFromCookie(cookie.Value)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, core.ErrRefreshTokenNotFound) {
			NewErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		NewErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + tokens[0],
		HttpOnly: true,
		MaxAge:   3600,
		Path:     "/",
	})

	NewTokenResponse(w, http.StatusOK, tokens[1])
}

func getCodeFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	code := vars["code"]
	if code == "" {
		return "", core.ErrEmptyCode
	}

	return code, nil
}

func getTokenFromCookie(cookieValue string) (string, error) {
	headerParts := strings.Split(cookieValue, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", core.ErrInvalidHeader
	}

	if len(headerParts[1]) == 0 {
		return "", core.ErrEmptyHeader
	}

	return headerParts[1], nil
}
