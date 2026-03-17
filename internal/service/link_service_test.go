package service

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"linkpulse/internal/cache"
	"linkpulse/internal/models"
	"linkpulse/internal/repository"
	"linkpulse/internal/ws"
)

func newSvc() *LinkService {
	repo := repository.NewMemoryRepo()
	return NewLinkService(repo, repo, cache.NewMemory(), ws.NewHub(), "http://localhost:8080", "test-cookie-secret")
}

func TestCreateAndResolve(t *testing.T) {
	svc := newSvc()
	l, short, err := svc.Create(context.Background(), models.CreateLinkInput{LongURL: "https://example.com", CustomAlias: "alias1"})
	if err != nil {
		t.Fatal(err)
	}
	if short == "" || l.ShortCode != "alias1" {
		t.Fatal("expected alias short link")
	}
	r, err := svc.Resolve(context.Background(), l.ShortCode)
	if err != nil || r.LongURL != "https://example.com" {
		t.Fatal("resolve failed")
	}
}

func TestValidatePassword(t *testing.T) {
	svc := newSvc()
	l, _, err := svc.Create(context.Background(), models.CreateLinkInput{LongURL: "https://example.com", CustomAlias: "alias2", Password: "topsecret"})
	if err != nil {
		t.Fatal(err)
	}
	if svc.ValidatePassword(l, "topsecret") != nil {
		t.Fatal("expected password success")
	}
	if svc.ValidatePassword(l, "bad") == nil {
		t.Fatal("expected password failure")
	}
}

func TestTrackClickAsync(t *testing.T) {
	svc := newSvc()
	l, _, _ := svc.Create(context.Background(), models.CreateLinkInput{LongURL: "https://example.com", CustomAlias: "alias3"})
	req := httptest.NewRequest("GET", "/alias3", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	svc.TrackClickAsync(l, req)
	time.Sleep(20 * time.Millisecond)
	s, err := svc.Summary(context.Background(), "alias3")
	if err != nil || s.TotalClicks == 0 {
		t.Fatal("expected click tracked")
	}
}

func TestDeleteInvalidatesCache(t *testing.T) {
	svc := newSvc()
	_, _, err := svc.Create(context.Background(), models.CreateLinkInput{LongURL: "https://example.com", CustomAlias: "alias4"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(context.Background(), "alias4"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Resolve(context.Background(), "alias4"); err == nil {
		t.Fatal("expected deleted link to be unresolved")
	}
}

func TestAccessTokenValidation(t *testing.T) {
	svc := newSvc()
	l, _, err := svc.Create(context.Background(), models.CreateLinkInput{LongURL: "https://example.com", CustomAlias: "alias5", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	token := svc.AccessToken(l)
	if token == "" {
		t.Fatal("expected signed token")
	}
	if !svc.ValidateAccessToken(l, token) {
		t.Fatal("expected token validation success")
	}
	if svc.ValidateAccessToken(l, "ok") {
		t.Fatal("forged cookie value must not pass")
	}
}
