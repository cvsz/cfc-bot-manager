package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/cvsz/cfc-bot-manager/backend/internal/messagestyle"
)

var roomIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

type roomStyleResponse struct {
	Style  messagestyle.Style `json:"style"`
	Source string              `json:"source"`
}

type apiError struct {
	Error string `json:"error"`
}

func NewRoomStyleHandler(store messagestyle.Store, token string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/message-style/default", func(w http.ResponseWriter, r *http.Request) {
		v, err := store.GetDefault(r.Context())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		writeAPIJSON(w, http.StatusOK, roomStyleResponse{v, "global_default"})
	})
	mux.HandleFunc("PUT /api/v1/message-style/default", func(w http.ResponseWriter, r *http.Request) {
		v, ok := decodeStyle(w, r)
		if !ok {
			return
		}
		if err := store.SaveDefault(r.Context(), v); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		writeAPIJSON(w, http.StatusOK, roomStyleResponse{v, "global_default"})
	})
	mux.HandleFunc("GET /api/v1/rooms/{roomID}/message-style", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("roomID")
		if !validRoomID(id) {
			writeAPIError(w, http.StatusBadRequest, "invalid_room_id")
			return
		}
		v, found, err := store.GetRoom(r.Context(), id)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		source := "room"
		if !found {
			v, err = store.GetDefault(r.Context())
			source = "global_default"
		}
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		writeAPIJSON(w, http.StatusOK, roomStyleResponse{v, source})
	})
	mux.HandleFunc("PUT /api/v1/rooms/{roomID}/message-style", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("roomID")
		if !validRoomID(id) {
			writeAPIError(w, http.StatusBadRequest, "invalid_room_id")
			return
		}
		v, ok := decodeStyle(w, r)
		if !ok {
			return
		}
		if err := store.SaveRoom(r.Context(), id, v); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		writeAPIJSON(w, http.StatusOK, roomStyleResponse{v, "room"})
	})
	mux.HandleFunc("DELETE /api/v1/rooms/{roomID}/message-style", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("roomID")
		if !validRoomID(id) {
			writeAPIError(w, http.StatusBadRequest, "invalid_room_id")
			return
		}
		if err := store.DeleteRoom(r.Context(), id); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		v, err := store.GetDefault(r.Context())
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		writeAPIJSON(w, http.StatusOK, roomStyleResponse{v, "global_default"})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authorized(r, token) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func authorized(r *http.Request, token string) bool {
	if token == "" {
		return false
	}
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	provided := strings.TrimPrefix(header, "Bearer ")
	if provided == "" || strings.ContainsAny(provided, " \t\r\n") {
		return false
	}
	expectedHash := sha256.Sum256([]byte(token))
	providedHash := sha256.Sum256([]byte(provided))
	return subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1
}

func validRoomID(value string) bool { return roomIDPattern.MatchString(value) }

func decodeStyle(w http.ResponseWriter, r *http.Request) (messagestyle.Style, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var style messagestyle.Style
	if err := decoder.Decode(&style); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeAPIError(w, http.StatusRequestEntityTooLarge, "request_too_large")
		} else {
			writeAPIError(w, http.StatusBadRequest, "invalid_json")
		}
		return messagestyle.Style{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeAPIError(w, http.StatusBadRequest, "invalid_json")
		return messagestyle.Style{}, false
	}
	normalized, err := style.Normalize()
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, "invalid_style")
		return messagestyle.Style{}, false
	}
	return normalized, true
}

func writeAPIJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, code string) {
	writeAPIJSON(w, status, apiError{code})
}
