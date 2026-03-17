package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"linkpulse/internal/cache"
	"linkpulse/internal/config"
	"linkpulse/internal/http/handlers"
	mid "linkpulse/internal/http/middleware"
	"linkpulse/internal/logger"
	"linkpulse/internal/metrics"
	"linkpulse/internal/repository"
	"linkpulse/internal/service"
	"linkpulse/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New()
	repo := repository.NewMemoryRepo()
	hub := ws.NewHub()
	svc := service.NewLinkService(repo, repo, cache.NewMemory(), hub, cfg.BaseURL, cfg.AccessCookieKey)
	h, err := handlers.New(svc, hub)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", mid.Metrics(h.Routes()))
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "links_created %d\nclicks_total %d\nstream_connections %d\n", metrics.LinksCreated.Load(), metrics.ClicksTotal.Load(), metrics.WSConnections.Load())
	})

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: mux, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout}
	go func() { log.Printf("server started on :%s", cfg.HTTPPort); _ = srv.ListenAndServe() }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
