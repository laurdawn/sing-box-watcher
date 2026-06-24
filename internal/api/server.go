package api

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/laurdawn/sing-box-watcher/internal/collector"
	"github.com/laurdawn/sing-box-watcher/internal/config"
	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type Server struct {
	cfg      *config.Config
	db       *sql.DB
	settings *store.Settings
	settingsMu sync.RWMutex
	manager  *collector.Manager
}

func NewServer(cfg *config.Config, db *sql.DB, settings *store.Settings, manager *collector.Manager) *Server {
	return &Server{cfg: cfg, db: db, settings: settings, manager: manager}
}

func (s *Server) Handler(static http.FileSystem) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/instances", s.handleInstances)
		r.Get("/traffic", s.handleTraffic)
		r.Get("/connections", s.handleConnections)
		r.Get("/connections/active", s.handleActiveConnections)
		r.Get("/connections/inbounds", s.handleInbounds)
		r.Get("/connections/outbounds", s.handleOutbounds)
		r.Get("/stats/top-domains", s.handleTopDomains)
		r.Get("/stats/top-outbounds", s.handleTopOutbounds)
		r.Get("/stats/source-regions", s.handleSourceRegions)
		r.Get("/stats/top-source-ips", s.handleTopSourceIPs)
		r.Get("/config", s.handleGetConfig)
		r.Put("/config", s.handleSaveConfig)
		r.Post("/geo/lookup", s.handleGeoLookup)
		r.Get("/service/info", s.handleServiceInfo)
		r.Get("/groups", s.handleGroups)
		r.Get("/groups/outbounds", s.handleGetOutbounds)
		r.Post("/groups/select", s.handleSelectOutbound)
		r.Post("/groups/urltest", s.handleURLTest)
	})

	r.Get("/ws/traffic", s.handleWsTraffic)
	r.Get("/ws/groups", s.handleWsGroups)
	r.Get("/ws/log", s.handleWsLog)

	if static != nil {
		r.Handle("/*", http.FileServer(static))
	}

	return r
}
