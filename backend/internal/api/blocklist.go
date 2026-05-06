package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type addSourceRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type domainRequest struct {
	Domain string `json:"domain"`
}

type toggleRequest struct {
	Enabled bool `json:"enabled"`
}

func (s *Server) handleGetBlocklistSources(w http.ResponseWriter, r *http.Request) {
	sources := s.blocklist.GetSources()
	jsonResponse(w, sources)
}

func (s *Server) handleAddBlocklistSource(w http.ResponseWriter, r *http.Request) {
	var req addSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.URL == "" {
		jsonError(w, "name and url are required", http.StatusBadRequest)
		return
	}

	source, err := s.blocklist.AddSource(req.Name, req.URL)
	if err != nil {
		jsonError(w, "failed to add source: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, source)
}

func (s *Server) handleDeleteBlocklistSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.blocklist.RemoveSource(id); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]string{"message": "source removed"})
}

func (s *Server) handleToggleBlocklistSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req toggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.blocklist.ToggleSource(id, req.Enabled); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]string{"message": "source toggled"})
}

func (s *Server) handleUpdateBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := s.blocklist.Update(); err != nil {
		jsonError(w, "failed to update blocklist: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "blocklist updated successfully"})
}

func (s *Server) handleGetBlocklistStats(w http.ResponseWriter, r *http.Request) {
	stats := s.blocklist.GetStats()
	jsonResponse(w, stats)
}

func (s *Server) handleBlockDomain(w http.ResponseWriter, r *http.Request) {
	var req domainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		jsonError(w, "domain is required", http.StatusBadRequest)
		return
	}
	s.blocklist.BlockDomain(req.Domain)
	jsonResponse(w, map[string]string{"message": "domain blocked"})
}

func (s *Server) handleUnblockDomain(w http.ResponseWriter, r *http.Request) {
	var req domainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		jsonError(w, "domain is required", http.StatusBadRequest)
		return
	}
	s.blocklist.UnblockDomain(req.Domain)
	jsonResponse(w, map[string]string{"message": "domain unblocked"})
}

func (s *Server) handleGetManualBlocks(w http.ResponseWriter, r *http.Request) {
	domains := s.blocklist.GetManualBlocks()
	jsonResponse(w, domains)
}

func (s *Server) handleGetWhitelist(w http.ResponseWriter, r *http.Request) {
	domains := s.blocklist.GetWhitelist()
	jsonResponse(w, domains)
}

func (s *Server) handleAddWhitelist(w http.ResponseWriter, r *http.Request) {
	var req domainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		jsonError(w, "domain is required", http.StatusBadRequest)
		return
	}
	s.blocklist.WhitelistDomain(req.Domain)
	jsonResponse(w, map[string]string{"message": "domain whitelisted"})
}

func (s *Server) handleRemoveWhitelist(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	if domain == "" {
		jsonError(w, "domain is required", http.StatusBadRequest)
		return
	}
	s.blocklist.RemoveWhitelist(domain)
	jsonResponse(w, map[string]string{"message": "domain removed from whitelist"})
}

func (s *Server) handleFlushCache(w http.ResponseWriter, r *http.Request) {
	if err := s.control.FlushCache(); err != nil {
		jsonError(w, "failed to flush cache: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "cache flushed successfully"})
}
