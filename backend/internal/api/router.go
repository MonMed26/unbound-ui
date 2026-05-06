package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/unbound-ui/backend/internal/auth"
	"github.com/unbound-ui/backend/internal/blocklist"
	"github.com/unbound-ui/backend/internal/unbound"
)

type Server struct {
	router       chi.Router
	auth         *auth.Service
	control      *unbound.Control
	configMgr    *unbound.ConfigManager
	blocklist    *blocklist.Manager
	frontendFS   embed.FS
	onAuthChange func()
}

type Config struct {
	Auth          *auth.Config
	UnboundConfig string
	ControlPath   string
	BlocklistDir  string
	BlocklistOut  string
	FrontendFS    embed.FS
}

func NewServer(cfg *Config) *Server {
	s := &Server{
		auth:       auth.NewService(cfg.Auth),
		control:    unbound.NewControl(cfg.ControlPath),
		configMgr:  unbound.NewConfigManager(cfg.UnboundConfig),
		blocklist:  blocklist.NewManager(cfg.BlocklistDir, cfg.BlocklistOut),
		frontendFS: cfg.FrontendFS,
	}

	s.setupRouter()
	return s
}

func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/auth/login", s.handleLogin)
	r.Post("/api/auth/setup", s.handleSetup)
	r.Get("/api/auth/status", s.handleAuthStatus)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(s.auth.Middleware)

		// Dashboard
		r.Get("/api/stats", s.handleGetStats)

		// Config
		r.Get("/api/config", s.handleGetConfig)
		r.Put("/api/config", s.handleUpdateConfig)
		r.Post("/api/config/validate", s.handleValidateConfig)
		r.Post("/api/config/reload", s.handleReload)

		// Zones
		r.Get("/api/zones", s.handleGetZones)
		r.Post("/api/zones", s.handleAddZone)
		r.Delete("/api/zones/{name}", s.handleDeleteZone)

		// Zone data
		r.Get("/api/zones/data", s.handleGetZoneData)
		r.Post("/api/zones/data", s.handleAddZoneData)
		r.Delete("/api/zones/data/{name}", s.handleDeleteZoneData)

		// Blocklist
		r.Get("/api/blocklist/sources", s.handleGetBlocklistSources)
		r.Post("/api/blocklist/sources", s.handleAddBlocklistSource)
		r.Delete("/api/blocklist/sources/{id}", s.handleDeleteBlocklistSource)
		r.Put("/api/blocklist/sources/{id}/toggle", s.handleToggleBlocklistSource)
		r.Post("/api/blocklist/update", s.handleUpdateBlocklist)
		r.Get("/api/blocklist/stats", s.handleGetBlocklistStats)

		// Blocklist domains
		r.Post("/api/blocklist/block", s.handleBlockDomain)
		r.Post("/api/blocklist/unblock", s.handleUnblockDomain)
		r.Get("/api/blocklist/manual", s.handleGetManualBlocks)

		// Whitelist
		r.Get("/api/blocklist/whitelist", s.handleGetWhitelist)
		r.Post("/api/blocklist/whitelist", s.handleAddWhitelist)
		r.Delete("/api/blocklist/whitelist/{domain}", s.handleRemoveWhitelist)

		// Cache
		r.Post("/api/cache/flush", s.handleFlushCache)
	})

	// Serve frontend SPA
	s.serveFrontend(r)

	s.router = r
}

func (s *Server) serveFrontend(r chi.Router) {
	// Try to serve embedded frontend files
	subFS, err := fs.Sub(s.frontendFS, "frontend/dist")
	if err != nil {
		// Frontend not embedded, skip
		return
	}

	fileServer := http.FileServer(http.FS(subFS))

	// Read index.html once for SPA fallback
	indexHTML, _ := fs.ReadFile(subFS, "index.html")

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists in the embedded FS
		f, err := subFS.Open(path[1:]) // Remove leading /
		if err != nil {
			// File not found, serve index.html for SPA routing
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(indexHTML)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) GetAuthConfig() *auth.Config {
	return s.auth.GetConfig()
}

func (s *Server) OnAuthChange(fn func()) {
	s.onAuthChange = fn
}

func (s *Server) notifyAuthChange() {
	if s.onAuthChange != nil {
		s.onAuthChange()
	}
}
