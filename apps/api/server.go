package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"orbit/apps/api/handler"
)

type Server struct {
	db      *pgxpool.Pool
	handler *handler.Handler
}

func NewServer(db *pgxpool.Pool, h *handler.Handler) *Server {
	return &Server{db: db, handler: h}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", s.health)
	r.Get("/ready", s.ready)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/users", s.handler.CreateUser)
		r.Route("/users/{userID}", func(r chi.Router) {
			r.Get("/", s.handler.GetUser)
			r.Put("/onboarding", s.handler.CompleteOnboarding)
		})
		r.Post("/conversations", s.handler.CreateConversation)
		r.Route("/conversations/{conversationID}", func(r chi.Router) {
			r.Get("/messages", s.handler.ListMessages)
			r.Post("/messages", s.handler.SendMessage)
		})
	})

	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
