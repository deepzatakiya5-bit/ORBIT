package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"

	"orbit/pkg/llm"
	"orbit/pkg/models"
	"orbit/pkg/store"
)

type createConversationRequest struct {
	Title *string `json:"title,omitempty"`
}

// GetConversation returns the authenticated user's single conversation thread.
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := authUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	conv, err := h.store.GetConversationByUser(r.Context(), userID)
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{"conversation": nil})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load conversation")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"conversation": conv})
}

// CreateConversation returns the user's existing conversation or creates one with a greeting.
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := authUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createConversationRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
	}

	user, err := h.store.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	if err := store.RequireOnboarded(user); err != nil {
		writeError(w, http.StatusForbidden, "complete onboarding before starting a conversation")
		return
	}

	existing, err := h.store.GetConversationByUser(r.Context(), userID)
	if err == nil {
		writeJSON(w, http.StatusOK, models.CreateConversationResponse{
			Conversation: existing,
		})
		return
	}
	if !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "failed to load conversation")
		return
	}

	conv, err := h.store.CreateConversation(r.Context(), userID, req.Title)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existing, getErr := h.store.GetConversationByUser(r.Context(), userID)
			if getErr != nil {
				writeError(w, http.StatusInternalServerError, "failed to load conversation")
				return
			}
			writeJSON(w, http.StatusOK, models.CreateConversationResponse{Conversation: existing})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	convID, err := parseUUID(conv.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid conversation id")
		return
	}

	greetingText, err := h.generateGreeting(r, user)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to generate greeting")
		return
	}

	greeting, err := h.store.CreateMessage(r.Context(), convID, "assistant", greetingText)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save greeting")
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateConversationResponse{
		Conversation: conv,
		Greeting:     &greeting,
	})
}

func (h *Handler) generateGreeting(r *http.Request, user models.User) (string, error) {
	if _, ok := h.llm.(*llm.Stub); ok {
		return llm.StubGreeting(user), nil
	}
	return h.llm.Chat(r.Context(), llm.GreetingPrompt(user))
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	convID, err := conversationIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	messages, err := h.store.ListMessages(r.Context(), convID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}
	if messages == nil {
		messages = []models.Message{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
}
