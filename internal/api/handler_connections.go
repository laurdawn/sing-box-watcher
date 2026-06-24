package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, to := parseTimeRange(r)

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	f := store.ConnectionFilter{
		Instance:     q.Get("instance"),
		Inbound:      q.Get("inbound"),
		InboundType:  q.Get("inbound_type"),
		Outbound:     q.Get("outbound"),
		Search:       q.Get("search"),
		SourceSearch: q.Get("source"),
		Rule:         q.Get("rule"),
		From:         from,
		To:           to,
		Page:         page,
		Limit:        limit,
		SortBy:       q.Get("sort_by"),
		SortDir:      q.Get("sort_dir"),
	}

	conns, total, err := store.QueryConnections(s.db, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if conns == nil {
		conns = []store.Connection{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":       total,
		"page":        f.Page,
		"limit":       f.Limit,
		"connections": conns,
	})
}

func (s *Server) handleActiveConnections(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ConnectionFilter{
		Instance:     q.Get("instance"),
		Search:       q.Get("search"),
		SourceSearch: q.Get("source"),
		ActiveOnly:   true,
		From:         0,
		To:           time.Now().UnixMilli(),
		Limit:        200,
		Page:         1,
	}
	conns, total, err := store.QueryConnections(s.db, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if conns == nil {
		conns = []store.Connection{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":       total,
		"connections": conns,
	})
}

func (s *Server) handleInbounds(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	inbounds, err := store.QueryInbounds(s.db, instance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if inbounds == nil {
		inbounds = []string{}
	}
	writeJSON(w, http.StatusOK, inbounds)
}

func (s *Server) handleOutbounds(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	outbounds, err := store.QueryOutbounds(s.db, instance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if outbounds == nil {
		outbounds = []string{}
	}
	writeJSON(w, http.StatusOK, outbounds)
}

func (s *Server) handleTopDomains(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours <= 0 {
		hours = 24
	}
	from := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()

	domains, err := store.QueryTopDomains(s.db, instance, from, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if domains == nil {
		domains = []store.TopDomain{}
	}
	writeJSON(w, http.StatusOK, domains)
}

func (s *Server) handleTopOutbounds(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours <= 0 {
		hours = 24
	}
	from := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()

	outbounds, err := store.QueryTopOutbounds(s.db, instance, from, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if outbounds == nil {
		outbounds = []store.TopOutbound{}
	}
	writeJSON(w, http.StatusOK, outbounds)
}
