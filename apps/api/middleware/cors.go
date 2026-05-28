package middleware

import (
	"net/http"
	"os"
	"strings"
)

// DevCORS allows the local FE dev server (make fe) to call the API from another origin.
func DevCORS(next http.Handler) http.Handler {
	allowed := devOrigins()
	if len(allowed) == 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && originAllowed(origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func devOrigins() []string {
	if os.Getenv("FE_DEV") != "1" {
		return nil
	}
	raw := os.Getenv("FE_ORIGIN")
	if raw == "" {
		raw = "http://localhost:3000,http://127.0.0.1:3000"
	}
	var out []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			out = append(out, o)
		}
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if origin == a {
			return true
		}
	}
	return false
}
