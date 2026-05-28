package handler

import (
	"net/http"

	"github.com/google/uuid"

	"orbit/pkg/auth"
)

func authUserID(r *http.Request) (uuid.UUID, bool) {
	return auth.UserIDFromContext(r.Context())
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
