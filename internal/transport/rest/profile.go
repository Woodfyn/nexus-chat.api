package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Woodfyn/chat-api-backend-go/internal/core"

	"github.com/gorilla/mux"
)

func (h *Handler) initProfileRouter(api *mux.Router) {
	profile := api.PathPrefix("/profile").Subrouter()
	{
		profile.Use(h.AuthMiddleware)

		profile.HandleFunc("/", h.profileGet).Methods(http.MethodGet)
		profile.HandleFunc("/update", h.profileUpdate).Methods(http.MethodPatch)

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
// @Description get profile
// @ID getProfile
// @Produce json
// @Param input body core.GetProfileResp true "profile"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/profile/ [get]
func (h *Handler) profileGet(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	profile, err := h.profileService.GetProfile(r.Context(), userId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, profile)
}

// @Summary UpdateProfile
// @Tags Profile
// @Security ApiKeyAuth
// @Description update profile
// @ID updateProfile
// @Accept json
// @Produce json
// @Param input body core.UpdateUserReq true "user data to update"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/profile/update [patch]
func (h *Handler) profileUpdate(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req *core.UpdateUserReq
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer r.Body.Close()

	if err = h.profileService.UpdateProfile(r.Context(), &core.User{
		ID:       userId,
		Phone:    req.Phone,
		Username: req.Username,
	}); err != nil {
		if errors.Is(err, core.ErrDuplicatedKey) {
			h.newErrorResponse(w, http.StatusBadRequest, core.ErrThisCredIsAlready.Error())
			return
		}

		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary GetAvatars
// @Tags Profile
// @Security ApiKeyAuth
// @Description get all avatars
// @ID getAvatars
// @Produce json
// @Success 200 {array} core.GetAllUserAvatarsResp
// @Failure 400,500 {object} errorResponse
// @Router /api/profile/avatars [get]
func (h *Handler) profileAvatarGetAll(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	response, err := h.profileService.GetAvatars(r.Context(), userId)
	if err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, response)
}

// @Summary UploadAvatar
// @Tags Profile
// @Security ApiKeyAuth
// @Description upload avatar
// @ID uploadAvatar
// @Produce json
// @Param avatar formData file true "avatar"
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/profile/avatar/upload [post]
func (h *Handler) profileAvatarUpload(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	defer file.Close()

	if err := h.profileService.UploadAvatar(r.Context(), file, userId); err != nil {
		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

// @Summary DeleteAvatar
// @Tags Profile
// @Security ApiKeyAuth
// @Description delete avatar by avatar id
// @ID deleteAvatar
// @Param avatarId path string true "avatarId"
// @Produce json
// @Success 200
// @Failure 400,500 {object} errorResponse
// @Router /api/profile/avatar/delete/{avatarId} [delete]
func (h *Handler) profileAvatarDelete(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(int)
	if !ok {
		h.newErrorResponse(w, http.StatusBadRequest, core.ErrEmptyUserID.Error())
		return
	}

	avatarId, err := getAvatarIdFromRequest(r)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	avatarIdInt, err := strconv.Atoi(avatarId)
	if err != nil {
		h.newErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.profileService.DeleteAvatar(r.Context(), userId, avatarIdInt); err != nil {
		if errors.Is(err, core.ErrAvatarNotFound) {
			h.newErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		h.newErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.newResponse(w, http.StatusOK, nil)
}

func getAvatarIdFromRequest(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	avatarId := vars["avatarId"]
	if avatarId == "" {
		return "", core.ErrEmptyAvatarID
	}

	return avatarId, nil
}
