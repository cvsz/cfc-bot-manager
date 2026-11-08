package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthAndReadiness(t *testing.T) {
	handler := NewHandler("test", nil)
	for path, expected := range map[string]string{
		"/healthz": `"status":"ok"`,
		"/readyz": `"status":"ready"`,
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("%s: status=%d body=%q", path, recorder.Code, recorder.Body.String())
		}
		if recorder.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: missing no-store header", path)
		}
	}
}

func TestReadinessFailureReturns503WithoutLeakingError(t *testing.T) {
	handler := NewHandler("test", func(context.Context) error {
		return errors.New("database credential must not leak")
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if strings.Contains(recorder.Body.String(), "credential") {
		t.Fatalf("readiness response leaked internal error: %q", recorder.Body.String())
	}
}

func TestMutatingMethodRejected(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler("test", nil).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("unexpected response: status=%d allow=%q", recorder.Code, recorder.Header().Get("Allow"))
	}
}

func TestMetadata(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler("v0.1.0", nil).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"version":"v0.1.0"`) {
		t.Fatalf("unexpected metadata response: %d %q", recorder.Code, recorder.Body.String())
	}
}
