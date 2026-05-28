package handler

import (
	"net/http"

	"orbit/pkg/models"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.CreateUser(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, ok := h.issueToken(w, user.ID)
	if !ok {
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateUserResponse{
		User:        user,
		AccessToken: token,
	})
}
