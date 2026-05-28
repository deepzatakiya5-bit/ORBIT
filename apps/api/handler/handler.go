package handler

import (
	"encoding/json"
	"net/http"

	"orbit/pkg/auth"
	"orbit/pkg/llm"
	"orbit/pkg/store"
)

type Handler struct {
	store  *store.Store
	llm    llm.Provider
	tokens *auth.TokenService
}

func New(s *store.Store, provider llm.Provider, tokens *auth.TokenService) *Handler {
	return &Handler{store: s, llm: provider, tokens: tokens}
}

func (h *Handler) Store() *store.Store {
	return h.store
}

func (h *Handler) issueToken(w http.ResponseWriter, userID string) (string, bool) {
	id, err := parseUUID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return "", false
	}
	token, err := h.tokens.Issue(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return "", false
	}
	return token, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
