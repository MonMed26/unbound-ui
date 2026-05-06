package api

import (
	"net/http"

	"github.com/unbound-ui/backend/internal/unbound"
)

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	rawStats, err := s.control.Stats()
	if err != nil {
		jsonError(w, "failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stats := unbound.ParseStatistics(rawStats)
	jsonResponse(w, stats)
}
