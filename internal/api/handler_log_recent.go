package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/laurdawn/sing-box-watcher/internal/collector"
)

func (s *Server) handleRecentLogs(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	level := strings.ToUpper(r.URL.Query().Get("level"))
	keyword := r.URL.Query().Get("q")
	n := 100
	if v := r.URL.Query().Get("n"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 500 {
			n = parsed
		}
	}
	entries := s.manager.RecentLogs(instance, n, level, keyword)
	if entries == nil {
		entries = []collector.LogEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}
