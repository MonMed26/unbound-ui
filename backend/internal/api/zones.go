package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type addZoneRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type addZoneDataRequest struct {
	Data string `json:"data"`
}

func (s *Server) handleGetZones(w http.ResponseWriter, r *http.Request) {
	zones, err := s.control.ListLocalZones()
	if err != nil {
		jsonError(w, "failed to list zones: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, zones)
}

func (s *Server) handleAddZone(w http.ResponseWriter, r *http.Request) {
	var req addZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Type == "" {
		jsonError(w, "name and type are required", http.StatusBadRequest)
		return
	}

	if err := s.control.LocalZoneAdd(req.Name, req.Type); err != nil {
		jsonError(w, "failed to add zone: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "zone added successfully"})
}

func (s *Server) handleDeleteZone(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		jsonError(w, "zone name is required", http.StatusBadRequest)
		return
	}

	if err := s.control.LocalZoneRemove(name); err != nil {
		jsonError(w, "failed to remove zone: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "zone removed successfully"})
}

func (s *Server) handleGetZoneData(w http.ResponseWriter, r *http.Request) {
	data, err := s.control.ListLocalData()
	if err != nil {
		jsonError(w, "failed to list zone data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, data)
}

func (s *Server) handleAddZoneData(w http.ResponseWriter, r *http.Request) {
	var req addZoneDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Data == "" {
		jsonError(w, "data is required", http.StatusBadRequest)
		return
	}

	if err := s.control.LocalDataAdd(req.Data); err != nil {
		jsonError(w, "failed to add zone data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "zone data added successfully"})
}

func (s *Server) handleDeleteZoneData(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := s.control.LocalDataRemove(name); err != nil {
		jsonError(w, "failed to remove zone data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "zone data removed successfully"})
}
