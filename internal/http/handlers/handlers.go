package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"linkpulse/internal/metrics"
	"linkpulse/internal/models"
	"linkpulse/internal/service"
	"linkpulse/internal/ws"
)

type Handler struct {
	svc  *service.LinkService
	tmpl *template.Template
	hub  *ws.Hub
}

func New(svc *service.LinkService, hub *ws.Hub) (*Handler, error) {
	t, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc, tmpl: t, hub: hub}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", h.CreateForm)
	mux.HandleFunc("GET /links/", h.LinkDetails)
	mux.HandleFunc("GET /stream/", h.Stream)
	mux.HandleFunc("POST /access/", h.PasswordSubmit)
	mux.HandleFunc("POST /api/links", h.CreateLink)
	mux.HandleFunc("GET /api/links", h.ListLinks)
	mux.HandleFunc("GET /api/links/", h.linkAPIRouter)
	mux.HandleFunc("DELETE /api/links/", h.DeleteLink)
	mux.HandleFunc("/", h.Redirect)
	return mux
}

func codeFrom(path, prefix string) string { return strings.TrimPrefix(path, prefix) }

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	links, _ := h.svc.List(r.Context())
	_ = h.tmpl.ExecuteTemplate(w, "dashboard.html", map[string]any{"Links": links})
}
func (h *Handler) LinkDetails(w http.ResponseWriter, r *http.Request) {
	s, _ := h.svc.Summary(r.Context(), codeFrom(r.URL.Path, "/links/"))
	_ = h.tmpl.ExecuteTemplate(w, "details.html", s)
}

func (h *Handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	var exp *time.Time
	if v := r.FormValue("expires_at"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			exp = &t
		}
	}
	_, short, err := h.svc.Create(r.Context(), models.CreateLinkInput{LongURL: r.FormValue("long_url"), CustomAlias: r.FormValue("custom_alias"), ExpiresAt: exp, Password: r.FormValue("password")})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_, _ = w.Write([]byte("<div>Created: <a href='" + short + "'>" + short + "</a></div>"))
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var in models.CreateLinkInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad request", 400)
		return
	}
	link, short, err := h.svc.Create(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"link": link, "short_url": short})
}
func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	links, _ := h.svc.List(r.Context())
	_ = json.NewEncoder(w).Encode(links)
}
func (h *Handler) linkAPIRouter(w http.ResponseWriter, r *http.Request) {
	p := codeFrom(r.URL.Path, "/api/links/")
	if strings.HasSuffix(p, "/clicks") {
		h.GetSummary(w, r)
		return
	}
	h.GetLink(w, r)
}
func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request) {
	link, err := h.svc.Resolve(r.Context(), codeFrom(r.URL.Path, "/api/links/"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = json.NewEncoder(w).Encode(link)
}
func (h *Handler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	if h.svc.Delete(r.Context(), codeFrom(r.URL.Path, "/api/links/")) != nil {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(204)
}
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(codeFrom(r.URL.Path, "/api/links/"), "/clicks")
	s, err := h.svc.Summary(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = json.NewEncoder(w).Encode(s)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h.Dashboard(w, r)
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/")
	link, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		http.Error(w, "link expired", 410)
		return
	}
	if link.PasswordHash != nil {
		if c, _ := r.Cookie("lp_access_" + code); c == nil || c.Value != "ok" {
			_ = h.tmpl.ExecuteTemplate(w, "password.html", map[string]any{"Code": code})
			return
		}
	}
	h.svc.TrackClickAsync(link, r)
	http.Redirect(w, r, link.LongURL, http.StatusFound)
}

func (h *Handler) PasswordSubmit(w http.ResponseWriter, r *http.Request) {
	code := codeFrom(r.URL.Path, "/access/")
	link, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if h.svc.ValidatePassword(link, r.FormValue("password")) != nil {
		http.Error(w, "invalid password", 401)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "lp_access_" + code, Value: "ok", Path: "/", HttpOnly: true, MaxAge: 86400})
	http.Redirect(w, r, "/"+code, http.StatusFound)
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	code := codeFrom(r.URL.Path, "/stream/")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", 500)
		return
	}
	ch := h.hub.Subscribe(code)
	metrics.WSConnections.Add(1)
	defer func() { h.hub.Unsubscribe(code, ch); metrics.WSConnections.Add(-1) }()
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			_, _ = w.Write([]byte("data: " + string(msg) + "\n\n"))
			f.Flush()
		}
	}
}
