package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var res healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if res.Status != "ok" {
		t.Fatalf("expected status field to be 'ok', got %q", res.Status)
	}

	if res.Timestamp == "" {
		t.Fatal("expected timestamp to be present")
	}
}
