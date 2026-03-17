package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"linkpulse/internal/cache"
	"linkpulse/internal/repository"
	"linkpulse/internal/service"
	"linkpulse/internal/ws"
)

func TestLinkDetailsReturnsNotFoundForMissingCode(t *testing.T) {
	repo := repository.NewMemoryRepo()
	svc := service.NewLinkService(repo, repo, cache.NewMemory(), ws.NewHub(), "http://localhost:8080", "test-cookie-secret")
	h := &Handler{svc: svc}

	r := httptest.NewRequest(http.MethodGet, "/links/missing-code", nil)
	w := httptest.NewRecorder()
	h.LinkDetails(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
