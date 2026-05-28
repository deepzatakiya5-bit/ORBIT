package handler

import (
	"errors"
	"net/http"
	"strings"

	"orbit/pkg/models"
	"orbit/pkg/store"
)

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := bearerTokenFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := h.tokens.Parse(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	newToken, ok := h.issueToken(w, userID.String())
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{
		AccessToken: newToken,
		UserID:      userID.String(),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	email, phone, deviceID, err := normalizeLoginInput(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.store.FindUserByLogin(r.Context(), email, phone, deviceID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to find user")
		return
	}

	if err := store.RequireOnboarded(user); err != nil {
		writeError(w, http.StatusForbidden, "onboarding not completed")
		return
	}

	token, ok := h.issueToken(w, user.ID)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, models.LoginResponse{
		User:        user,
		AccessToken: token,
	})
}

func normalizeLoginInput(in models.LoginInput) (email, phone, deviceID string, err error) {
	email = strings.TrimSpace(strings.ToLower(in.Email))
	phone = strings.TrimSpace(in.Phone)
	deviceID = strings.TrimSpace(in.DeviceID)

	set := 0
	if email != "" {
		set++
	}
	if phone != "" {
		set++
	}
	if deviceID != "" {
		set++
	}
	if set != 1 {
		return "", "", "", errValidation("provide exactly one of: email, phone, or device_id")
	}
	return email, phone, deviceID, nil
}

func bearerTokenFromRequest(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", errValidation("missing or invalid authorization header")
	}
	token := strings.TrimSpace(strings.TrimPrefix(h, prefix))
	if token == "" {
		return "", errValidation("missing or invalid authorization header")
	}
	return token, nil
}
