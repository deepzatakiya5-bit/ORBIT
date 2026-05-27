package handler

import (
	"encoding/json"
	"net/http"

	"orbit/pkg/llm"
	"orbit/pkg/store"
)

type Handler struct {
	store *store.Store
	llm   llm.Provider
}

func New(s *store.Store, provider llm.Provider) *Handler {
	return &Handler{store: s, llm: provider}
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
