package api

import (
	"encoding/json"
	"net/http"

	"github.com/laurdawn/sing-box-watcher/internal/geo"
)

func (s *Server) handleGeoLookup(w http.ResponseWriter, r *http.Request) {
	var ips []string
	if err := json.NewDecoder(r.Body).Decode(&ips); err != nil || len(ips) == 0 {
		writeJSON(w, http.StatusOK, map[string]geo.Info{})
		return
	}
	// 最多一次查 200 个
	if len(ips) > 200 {
		ips = ips[:200]
	}
	result := make(map[string]geo.Info, len(ips))
	for _, ip := range ips {
		if ip == "" {
			continue
		}
		info := geo.Lookup(ip)
		if info.CountryCode != "" {
			result[ip] = info
		}
	}
	writeJSON(w, http.StatusOK, result)
}
