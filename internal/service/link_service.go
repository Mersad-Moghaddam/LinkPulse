package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"linkpulse/internal/cache"
	"linkpulse/internal/metrics"
	"linkpulse/internal/models"
	"linkpulse/internal/repository"
	"linkpulse/internal/ws"
)

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{4,20}$`)

type LinkService struct {
	links        repository.LinkRepository
	clicks       repository.ClickRepository
	cache        *cache.MemoryCache
	hub          *ws.Hub
	base         string
	cookieSecret []byte
}

func NewLinkService(links repository.LinkRepository, clicks repository.ClickRepository, cache *cache.MemoryCache, hub *ws.Hub, baseURL string, cookieSecret string) *LinkService {
	if cookieSecret == "" {
		cookieSecret = "linkpulse-dev-cookie-secret"
	}
	return &LinkService{links: links, clicks: clicks, cache: cache, hub: hub, base: strings.TrimRight(baseURL, "/"), cookieSecret: []byte(cookieSecret)}
}

func (s *LinkService) Create(ctx context.Context, in models.CreateLinkInput) (models.Link, string, error) {
	parsed, err := url.ParseRequestURI(in.LongURL)
	if err != nil || parsed.Host == "" || parsed.Hostname() == "" {
		return models.Link{}, "", errors.New("invalid long URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return models.Link{}, "", errors.New("long URL must use http or https")
	}
	if in.ExpiresAt != nil && in.ExpiresAt.Before(time.Now()) {
		return models.Link{}, "", errors.New("expiration must be in future")
	}
	code := in.CustomAlias
	if code == "" {
		code = randomCode(7)
	} else if !aliasRegex.MatchString(code) {
		return models.Link{}, "", errors.New("invalid alias")
	}
	var hash *string
	if in.Password != "" {
		h := hashPassword(in.Password)
		hash = &h
	}
	link := models.Link{ID: randomCode(16), ShortCode: code, LongURL: in.LongURL, PasswordHash: hash, ExpiresAt: in.ExpiresAt}
	created, err := s.links.Create(ctx, link)
	if err != nil {
		return models.Link{}, "", err
	}
	s.cache.SetLink(ctx, created)
	metrics.LinksCreated.Add(1)
	return created, s.base + "/" + created.ShortCode, nil
}

func (s *LinkService) Resolve(ctx context.Context, code string) (models.Link, error) {
	if l, ok := s.cache.GetLink(ctx, code); ok {
		return l, nil
	}
	l, err := s.links.GetByCode(ctx, code)
	if err != nil {
		return models.Link{}, err
	}
	s.cache.SetLink(ctx, l)
	return l, nil
}

func (s *LinkService) ValidatePassword(link models.Link, provided string) error {
	if link.PasswordHash == nil {
		return nil
	}
	if hashPassword(provided) != *link.PasswordHash {
		return repository.ErrUnauthorized
	}
	return nil
}

func (s *LinkService) TrackClickAsync(link models.Link, r *http.Request) {
	go func() {
		ua := r.UserAgent()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip == "" {
			ip = r.RemoteAddr
		}
		browser, os := parseUA(ua)
		click := models.Click{LinkID: link.ShortCode, Referrer: r.Referer(), UserAgent: ua, IP: ip, Browser: browser, OS: os, ClickedAt: time.Now().UTC()}
		_ = s.clicks.Record(context.Background(), click)
		s.cache.IncrCounters(context.Background(), link.ShortCode)
		metrics.ClicksTotal.Add(1)
		b, _ := json.Marshal(map[string]any{"code": link.ShortCode, "clicked_at": click.ClickedAt, "referrer": click.Referrer, "browser": click.Browser})
		s.hub.Broadcast(link.ShortCode, b)
	}()
}

func (s *LinkService) List(ctx context.Context) ([]models.Link, error) { return s.links.List(ctx) }
func (s *LinkService) Delete(ctx context.Context, code string) error {
	if err := s.links.DeleteByCode(ctx, code); err != nil {
		return err
	}
	s.cache.DeleteLink(ctx, code)
	return nil
}
func (s *LinkService) Summary(ctx context.Context, code string) (models.AnalyticsSummary, error) {
	return s.clicks.SummaryByCode(ctx, code)
}

func (s *LinkService) AccessToken(link models.Link) string {
	if link.PasswordHash == nil {
		return ""
	}
	mac := hmac.New(sha256.New, s.cookieSecret)
	_, _ = mac.Write([]byte(link.ShortCode + "|" + *link.PasswordHash))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *LinkService) ValidateAccessToken(link models.Link, token string) bool {
	if link.PasswordHash == nil || token == "" {
		return false
	}
	expected := s.AccessToken(link)
	return hmac.Equal([]byte(token), []byte(expected))
}

func randomCode(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:length]
}
func hashPassword(v string) string { s := sha256.Sum256([]byte(v)); return hex.EncodeToString(s[:]) }
func parseUA(ua string) (string, string) {
	browser, os := "other", "other"
	if strings.Contains(ua, "Chrome") {
		browser = "chrome"
	}
	if strings.Contains(ua, "Firefox") {
		browser = "firefox"
	}
	if strings.Contains(ua, "Windows") {
		os = "windows"
	}
	if strings.Contains(ua, "Linux") {
		os = "linux"
	}
	if strings.Contains(ua, "Mac") {
		os = "macos"
	}
	return browser, os
}
