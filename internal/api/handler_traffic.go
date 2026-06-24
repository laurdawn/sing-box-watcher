package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

func (s *Server) handleTraffic(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	from, to := parseTimeRange(r)

	points, err := store.QueryTraffic(s.db, instance, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if points == nil {
		points = []store.TrafficPoint{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"instance": instance,
		"points":   points,
	})
}

func parseTimeRange(r *http.Request) (from, to int64) {
	now := time.Now().UnixMilli()
	to = now
	from = now - 3600_000 // 默认 1 小时（毫秒）

	if v := r.URL.Query().Get("from"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			from = n
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			to = n
		}
	}
	return
}
