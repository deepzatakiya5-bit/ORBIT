package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"orbit/pkg/auth"
	"orbit/pkg/store"
)

func RequireAuth(tokens *auth.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, err.Error())
				return
			}

			userID, err := tokens.Parse(token)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePathUser ensures URL {userID} matches the authenticated user.
func RequirePathUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			writeAuthError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		pathID, err := uuid.Parse(chi.URLParam(r, "userID"))
		if err != nil {
			writeAuthError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		if authID != pathID {
			writeAuthError(w, http.StatusForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireConversationOwner ensures the conversation belongs to the authenticated user.
func RequireConversationOwner(st *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			convID, err := uuid.Parse(chi.URLParam(r, "conversationID"))
			if err != nil {
				writeAuthError(w, http.StatusBadRequest, "invalid conversation id")
				return
			}

			conv, err := st.GetConversation(r.Context(), convID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					writeAuthError(w, http.StatusNotFound, "conversation not found")
					return
				}
				writeAuthError(w, http.StatusInternalServerError, "failed to load conversation")
				return
			}

			convUserID, err := uuid.Parse(conv.UserID)
			if err != nil || authID != convUserID {
				writeAuthError(w, http.StatusForbidden, "forbidden")
				return
			}

			ctx := context.WithValue(r.Context(), conversationKey{}, conv)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type conversationKey struct{}

func bearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errMissingToken{}
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", errMissingToken{}
	}
	token := strings.TrimSpace(strings.TrimPrefix(h, prefix))
	if token == "" {
		return "", errMissingToken{}
	}
	return token, nil
}

type errMissingToken struct{}

func (errMissingToken) Error() string { return "missing or invalid authorization header" }

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
