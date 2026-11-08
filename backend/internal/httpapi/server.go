package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
)

type ReadinessCheck func(context.Context) error

type response struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

func NewHandler(version string, readiness ReadinessCheck) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", getOnly(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, response{Status: "ok"})
	}))
	mux.HandleFunc("/readyz", getOnly(func(w http.ResponseWriter, r *http.Request) {
		if readiness != nil {
			if err := readiness(r.Context()); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, response{Status: "not_ready"})
				return
			}
		}
		writeJSON(w, http.StatusOK, response{Status: "ready"})
	}))
	mux.HandleFunc("/api/v1/meta", getOnly(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, response{Status: "ok", Version: version})
	}))
	return securityHeaders(mux)
}

func getOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSON(w, http.StatusMethodNotAllowed, response{Status: "method_not_allowed"})
			return
		}
		next(w, r)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
