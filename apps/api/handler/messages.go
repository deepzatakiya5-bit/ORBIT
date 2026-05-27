package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"orbit/pkg/llm"
	"orbit/pkg/models"
	"orbit/pkg/store"
)

type sendMessageRequest struct {
	Content string `json:"content"`
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	convID, err := conversationIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	var req sendMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	conv, err := h.store.GetConversation(r.Context(), convID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "conversation not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load conversation")
		return
	}

	userID, err := uuid.Parse(conv.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	user, err := h.store.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	if err := store.RequireOnboarded(user); err != nil {
		writeError(w, http.StatusForbidden, "complete onboarding before chatting")
		return
	}

	userMsg, err := h.store.CreateMessage(r.Context(), convID, "user", req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save message")
		return
	}

	history, err := h.store.ListMessages(r.Context(), convID, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load history")
		return
	}

	llmMessages := make([]llm.Message, len(history))
	for i, m := range history {
		llmMessages[i] = llm.Message{Role: m.Role, Content: m.Content}
	}

	reply, err := h.llm.Chat(r.Context(), llm.WithSystemForUser(user, llmMessages))
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to generate reply")
		return
	}

	assistantMsg, err := h.store.CreateMessage(r.Context(), convID, "assistant", reply)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save reply")
		return
	}

	writeJSON(w, http.StatusOK, models.SendMessageResponse{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
	})
}

func conversationIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "conversationID"))
}
