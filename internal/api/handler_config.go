package api

import (
	"encoding/json"
	"net/http"

	"github.com/laurdawn/sing-box-watcher/internal/geo"
	"github.com/laurdawn/sing-box-watcher/internal/store"
)

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	s.settingsMu.RLock()
	defer s.settingsMu.RUnlock()
	writeJSON(w, http.StatusOK, s.settings)
}

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var body store.Settings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.RetentionDays <= 0 {
		body.RetentionDays = 7
	}
	for _, inst := range body.Instances {
		if inst.Name == "" || inst.API == "" {
			http.Error(w, "instance name and api are required", http.StatusBadRequest)
			return
		}
	}

	s.settingsMu.Lock()
	*s.settings = body
	s.settingsMu.Unlock()

	if err := store.SaveSettings(s.db, &body); err != nil {
		http.Error(w, "save settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// GeoIP 路径变更时热重载
	if body.GeoDBPath != "" {
		if err := geo.EnsureDB(body.GeoDBPath, body.GeoDBURL); err == nil {
			geo.Reinit(body.GeoDBPath)
		}
	}

	s.manager.Reload(body.Instances)

	// 热重载 MCP 开关
	if body.MCPEnabled {
		s.mcpGate.enable("http://" + s.cfg.Listen)
	} else {
		s.mcpGate.disable()
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
