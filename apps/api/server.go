package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"orbit/apps/api/handler"
	"orbit/apps/api/middleware"
	"orbit/pkg/auth"
)

type Server struct {
	db      *pgxpool.Pool
	handler *handler.Handler
	tokens  *auth.TokenService
}

func NewServer(db *pgxpool.Pool, h *handler.Handler, tokens *auth.TokenService) *Server {
	return &Server{db: db, handler: h, tokens: tokens}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.DevCORS)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})
	r.Handle("/ui/*", http.StripPrefix("/ui", webHandler()))

	r.Get("/health", s.health)
	r.Get("/ready", s.ready)

	r.Route("/v1", func(r chi.Router) {
		r.Post("/users", s.handler.CreateUser)
		r.Post("/auth/login", s.handler.Login)
		r.Post("/auth/refresh", s.handler.RefreshToken)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(s.tokens))

			r.Route("/users/{userID}", func(r chi.Router) {
				r.Use(middleware.RequirePathUser)
				r.Get("/", s.handler.GetUser)
				r.Put("/onboarding", s.handler.CompleteOnboarding)
			})

			r.Get("/conversations", s.handler.GetConversation)
			r.Post("/conversations", s.handler.CreateConversation)

			r.Route("/conversations/{conversationID}", func(r chi.Router) {
				r.Use(middleware.RequireConversationOwner(s.handler.Store()))
				r.Get("/messages", s.handler.ListMessages)
				r.Post("/messages", s.handler.SendMessage)
			})
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
