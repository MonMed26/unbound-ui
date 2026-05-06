package api

import (
	"encoding/json"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := s.auth.Login(req.Username, req.Password)
	if err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	jsonResponse(w, loginResponse{Token: token})
}

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	if !s.auth.IsSetupRequired() {
		jsonError(w, "setup already completed", http.StatusBadRequest)
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		jsonError(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	if err := s.auth.Setup(req.Username, req.Password); err != nil {
		jsonError(w, "failed to setup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Persist auth config to disk
	s.notifyAuthChange()

	// Generate token for immediate login
	token, err := s.auth.Login(req.Username, req.Password)
	if err != nil {
		jsonError(w, "setup completed but login failed", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, loginResponse{Token: token})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]interface{}{
		"setup_required": s.auth.IsSetupRequired(),
	})
}
