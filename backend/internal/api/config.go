package api

import (
	"encoding/json"
	"net/http"
)

type updateConfigRequest struct {
	Raw string `json:"raw"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.configMgr.ReadConfig()
	if err != nil {
		jsonError(w, "failed to read config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, config)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Raw == "" {
		jsonError(w, "config content is required", http.StatusBadRequest)
		return
	}

	if err := s.configMgr.WriteConfig(req.Raw); err != nil {
		jsonError(w, "failed to write config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"success": true,
		"message": "config updated successfully",
	})
}

func (s *Server) handleValidateConfig(w http.ResponseWriter, r *http.Request) {
	err := s.configMgr.ValidateConfig()
	if err != nil {
		jsonResponse(w, map[string]interface{}{
			"valid":   false,
			"message": err.Error(),
		})
		return
	}
	jsonResponse(w, map[string]interface{}{
		"valid":   true,
		"message": "configuration is valid",
	})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if err := s.control.Reload(); err != nil {
		jsonError(w, "failed to reload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "unbound reloaded successfully"})
}
