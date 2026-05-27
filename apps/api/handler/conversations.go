package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"orbit/pkg/llm"
	"orbit/pkg/models"
	"orbit/pkg/store"
)

type createConversationRequest struct {
	UserID string  `json:"user_id"`
	Title  *string `json:"title,omitempty"`
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var req createConversationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
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

	conv, err := h.store.CreateConversation(r.Context(), userID, req.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	convID, err := uuid.Parse(conv.ID)
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

	if _, err := h.store.GetConversation(r.Context(), convID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "conversation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load conversation")
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
