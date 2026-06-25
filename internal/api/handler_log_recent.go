package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type logResponse struct {
	TS      int64  `json:"ts"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (s *Server) handleRecentLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	instance := q.Get("instance")
	level := strings.ToUpper(q.Get("level"))
	keyword := q.Get("q")

	fromStr := q.Get("from")
	toStr := q.Get("to")
	limitStr := q.Get("limit")
	nStr := q.Get("n") // legacy param

	var from, to int64
	if v, err := strconv.ParseInt(fromStr, 10, 64); err == nil {
		from = v
	}
	if v, err := strconv.ParseInt(toStr, 10, 64); err == nil {
		to = v
	}

	// historical query via DB when time range is specified
	if from > 0 || to > 0 {
		limit := 200
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
		rows, err := store.QueryLogs(s.db, store.LogFilter{
			Instance: instance,
			Level:    level,
			Keyword:  keyword,
			From:     from,
			To:       to,
			Limit:    limit,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := make([]logResponse, len(rows))
		for i, r := range rows {
			resp[i] = logResponse{TS: r.TSMillis, Level: r.Level, Message: r.Message}
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	// in-memory recent query
	n := 100
	if v, err := strconv.Atoi(nStr); err == nil && v > 0 && v <= 500 {
		n = v
	} else if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 500 {
		n = v
	}
	entries := s.manager.RecentLogs(instance, n, level, keyword)
	resp := make([]logResponse, len(entries))
	for i, e := range entries {
		resp[i] = logResponse{TS: e.Time.UnixMilli(), Level: e.Level, Message: e.Message}
	}
	writeJSON(w, http.StatusOK, resp)
}
