package api

import (
	"database/sql"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	internalauth "github.com/laurdawn/sing-box-watcher/internal/auth"
	"github.com/laurdawn/sing-box-watcher/internal/collector"
	"github.com/laurdawn/sing-box-watcher/internal/config"
	internalmcp "github.com/laurdawn/sing-box-watcher/internal/mcp"
	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type Server struct {
	cfg        *config.Config
	db         *sql.DB
	settings   *store.Settings
	settingsMu sync.RWMutex
	manager    *collector.Manager
	mcpGate    *mcpGate
	authStore  *internalauth.Store
}

// mcpGate holds a swappable MCP handler, returning 503 when disabled.
type mcpGate struct {
	mu      sync.RWMutex
	handler http.Handler
}

func (g *mcpGate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	h := g.handler
	g.mu.RUnlock()
	if h == nil {
		http.Error(w, "MCP is disabled", http.StatusServiceUnavailable)
		return
	}
	h.ServeHTTP(w, r)
}

func (g *mcpGate) enable(baseURL string) {
	h := internalmcp.NewHandler(baseURL)
	g.mu.Lock()
	g.handler = h
	g.mu.Unlock()
}

func (g *mcpGate) disable() {
	g.mu.Lock()
	g.handler = nil
	g.mu.Unlock()
}

func NewServer(cfg *config.Config, db *sql.DB, settings *store.Settings, manager *collector.Manager) *Server {
	s := &Server{
		cfg:       cfg,
		db:        db,
		settings:  settings,
		manager:   manager,
		mcpGate:   &mcpGate{},
		authStore: internalauth.NewStore(),
	}
	if settings.MCPEnabled {
		s.mcpGate.enable("http://" + cfg.Listen)
	}
	return s
}

func (s *Server) Handler(static http.FileSystem) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// public auth endpoints
	r.Post("/api/auth/login", s.handleLogin)

	// protected API routes
	r.Group(func(r chi.Router) {
		r.Use(internalauth.Middleware(s.authStore))
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
			r.Get("/logs/recent", s.handleRecentLogs)
			r.Get("/groups", s.handleGroups)
			r.Get("/groups/outbounds", s.handleGetOutbounds)
			r.Post("/groups/select", s.handleSelectOutbound)
			r.Post("/groups/urltest", s.handleURLTest)
			// auth management
			r.Post("/auth/logout", s.handleLogout)
			r.Get("/auth/me", s.handleMe)
			r.Post("/auth/password", s.handleChangePassword)
			r.Post("/auth/regenerate-mcp-token", s.handleRegenerateMCPToken)
		})
		r.Get("/ws/traffic", s.handleWsTraffic)
		r.Get("/ws/groups", s.handleWsGroups)
		r.Get("/ws/log", s.handleWsLog)
	})

	// MCP: Bearer token auth
	r.Group(func(r chi.Router) {
		r.Use(internalauth.MCPMiddleware(func() string {
			s.settingsMu.RLock()
			defer s.settingsMu.RUnlock()
			return s.settings.MCPToken
		}))
		r.Mount("/mcp", s.mcpGate)
	})

	if static != nil {
		r.Handle("/*", http.FileServer(static))
	}

	return r
}


