package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"orbit/pkg/models"
	"orbit/pkg/store"
)

var validGenders = map[string]struct{}{
	"male":              {},
	"female":            {},
	"non_binary":        {},
	"other":             {},
	"prefer_not_to_say": {},
}

var validCommunicationStyles = map[string]string{
	"warm_short":    "warm and concise",
	"warm_long":     "warm and detailed",
	"direct_short":  "direct and concise",
	"playful_short": "playful and concise",
}

var validWhyHere = map[string]string{
	"companionship":   "companionship and someone to talk to",
	"stress_support":  "support during stress or hard times",
	"goals":           "help with goals and accountability",
	"loneliness":      "feeling lonely or disconnected",
	"curiosity":       "curiosity about AI companions",
	"self_growth":     "self-reflection and personal growth",
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
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

	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var input models.OnboardingInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	birthdate, whyHere, err := validateOnboarding(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := h.store.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	if existing.OnboardingCompleted {
		writeError(w, http.StatusConflict, "onboarding already completed")
		return
	}

	user, err := h.store.CompleteOnboarding(r.Context(), userID, store.OnboardingProfile{
		Name:               strings.TrimSpace(input.Name),
		Nickname:           strings.TrimSpace(input.Nickname),
		Gender:             input.Gender,
		Birthdate:          birthdate,
		Occupation:         strings.TrimSpace(input.Occupation),
		PreferredLanguage:  strings.TrimSpace(input.PreferredLanguage),
		Timezone:           strings.TrimSpace(input.Timezone),
		CommunicationStyle: input.CommunicationStyle,
		WhyHere:            whyHere,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save profile")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func validateOnboarding(in models.OnboardingInput) (time.Time, []string, error) {
	if strings.TrimSpace(in.Name) == "" {
		return time.Time{}, nil, errValidation("name is required")
	}
	if _, ok := validGenders[in.Gender]; !ok {
		return time.Time{}, nil, errValidation("gender must be one of: male, female, non_binary, other, prefer_not_to_say")
	}
	if strings.TrimSpace(in.Occupation) == "" {
		return time.Time{}, nil, errValidation("occupation is required")
	}
	if strings.TrimSpace(in.PreferredLanguage) == "" {
		return time.Time{}, nil, errValidation("preferred_language is required")
	}
	if strings.TrimSpace(in.Timezone) == "" {
		return time.Time{}, nil, errValidation("timezone is required (IANA name, e.g. Asia/Kolkata)")
	}
	if _, err := time.LoadLocation(strings.TrimSpace(in.Timezone)); err != nil {
		return time.Time{}, nil, errValidation("timezone must be a valid IANA timezone (e.g. Asia/Kolkata, America/New_York)")
	}
	if _, ok := validCommunicationStyles[in.CommunicationStyle]; !ok {
		return time.Time{}, nil, errValidation("communication_style must be one of: warm_short, warm_long, direct_short, playful_short")
	}

	whyHere, err := normalizeWhyHere(in.WhyHere)
	if err != nil {
		return time.Time{}, nil, err
	}

	birthdate, err := time.Parse("2006-01-02", in.Birthdate)
	if err != nil {
		return time.Time{}, nil, errValidation("birthdate must be YYYY-MM-DD")
	}
	if birthdate.After(time.Now()) {
		return time.Time{}, nil, errValidation("birthdate cannot be in the future")
	}
	age := time.Now().Year() - birthdate.Year()
	if age < 13 {
		return time.Time{}, nil, errValidation("you must be at least 13 years old to use ORBIT")
	}

	return birthdate, whyHere, nil
}

func normalizeWhyHere(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, errValidation("why_here must include at least one reason")
	}
	if len(values) > 5 {
		return nil, errValidation("why_here can include at most 5 reasons")
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := validWhyHere[v]; !ok {
			return nil, errValidation("why_here values must be: companionship, stress_support, goals, loneliness, curiosity, self_growth")
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, errValidation("why_here must include at least one reason")
	}
	return out, nil
}

type validationError string

func (e validationError) Error() string { return string(e) }

func errValidation(msg string) error { return validationError(msg) }

func userIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "userID"))
}
