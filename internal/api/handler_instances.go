package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleInstances(w http.ResponseWriter, r *http.Request) {
	stats := s.manager.Stats()
	writeJSON(w, http.StatusOK, stats)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
