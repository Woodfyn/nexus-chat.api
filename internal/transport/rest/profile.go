package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (h *Handler) initProfileRouter(api *mux.Router) {
	profile := api.PathPrefix("/profile").Subrouter()
	{
		profile.Use(h.AuthMiddleware)

		profile.HandleFunc("/", h.profileGet).Methods(http.MethodGet)
		profile.HandleFunc("/update", h.profileUpdate).Methods(http.MethodPut)

		avatar := profile.PathPrefix("/avatar").Subrouter()
		{
			avatar.HandleFunc("/get", h.profileAvatarGetAll).Methods(http.MethodGet)
			avatar.HandleFunc("/upload", h.profileAvatarUpload).Methods(http.MethodPost)
			avatar.HandleFunc("/delete/{avatarId}", h.profileAvatarDelete).Methods(http.MethodDelete)
		}
	}
}

// @Summary GetProfile
// @Tags Profile
// @Security ApiKeyAuth
// @Description Get profile
// @ID getProfile
// @Produce  json
// @Param input body core.GetProfileResponse true "profile"
// @Success 200
// @Failure 400,500 {object} response
// @Router /api/profile/ [get]
func (h *Handler) profileGet(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileGet -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	profile, err := h.profileService.GetProfile(r.Context(), userId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileGet -> GetProfile"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(profile)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileGet -> Marshal"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err = w.Write(response); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileGet -> Write"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary UpdateProfile
// @Tags Profile
// @Security ApiKeyAuth
// @Description Update profile
// @ID updateProfile
// @Accept  json
// @Produce  json
// @Param input body core.UpdateUserReq true "user data to update"
// @Success 200 {object} core.User
// @Failure 400,500 {object} response
// @Router /api/profile/update [patch]
func (h *Handler) profileUpdate(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> ReadAll"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req *core.UpdateUserReq
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Unmarshal"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Validate"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.profileService.UpdateProfile(r.Context(), &core.User{
		ID:       userId,
		Phone:    req.Phone,
		Username: req.Username,
	})
	if err != nil {
		if errors.Is(err, core.ErrThisCredIsAlready) {
			h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> UpdateProfile"}).Error(err)
			NewResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> JoinRoom"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(user)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Marshal"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return

	}

	if _, err = w.Write(response); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileUpdate -> Write"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary GetAvatars
// @Tags Profile
// @Security ApiKeyAuth
// @Description Get all avatars
// @ID getAvatars
// @Produce  json
// @Success 200 {array} core.GetAllUSerAvatarsResponse
// @Failure 400,500 {object} response
// @Router /api/profile/avatars [get]
func (h *Handler) profileAvatarGetAll(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarGetAll -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	response, err := h.profileService.GetAvatars(r.Context(), userId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarGetAll -> GetAvatars"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	responseBytes, err := json.Marshal(response)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarGetAll -> Marshal"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err = w.Write(responseBytes); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarGetAll -> Write"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary UploadAvatar
// @Tags Profile
// @Security ApiKeyAuth
// @Description Upload avatar
// @ID uploadAvatar
// @Param avatar formData file true "avatar"
// @Produce  json
// @Success 200
// @Failure 400,500 {object} response
// @Router /api/profile/avatar/upload [post]
func (h *Handler) profileAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarUpload -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarUpload -> FormFile"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer file.Close()

	if err := h.profileService.UploadAvatar(r.Context(), file, userId); err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarUpload -> UploadAvatar"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, "OK")
}

// @Summary DeleteAvatar
// @Tags Profile
// @Security ApiKeyAuth
// @Description Delete avatar
// @ID deleteAvatar
// @Param avatarId path string true "avatarId"
// @Produce  json
// @Success 200
// @Failure 400,500 {object} response
// @Router /api/profile/avatar/delete/{avatarId} [delete]
func (h *Handler) profileAvatarDelete(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarDelete -> Context"}).Error(core.ErrEmptyUserId)
		NewResponse(w, http.StatusBadRequest, core.ErrEmptyUserId.Error())
		return
	}

	avatarId, err := getAvatarIdFromRequest(r)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarDelete -> getAvatarIdFromRequest"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	avatarIdInt, err := strconv.Atoi(avatarId)
	if err != nil {
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarDelete -> Atoi"}).Error(err)
		NewResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.profileService.DeleteAvatar(r.Context(), userId, avatarIdInt); err != nil {
		if errors.Is(err, core.ErrAvatarNotFound) {
			h.log.WithFields(logrus.Fields{"handler": "profileAvatarDelete -> DeleteAvatar"}).Error(err)
			NewResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		h.log.WithFields(logrus.Fields{"handler": "profileAvatarDelete -> DeleteAvatar"}).Error(err)
		NewResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	NewResponse(w, http.StatusOK, "OK")
}

func getAvatarIdFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	avatarId := vars["avatarId"]
	if avatarId == "" {
		return "", core.ErrEmptyAvatarId
	}

	return avatarId, nil
}
