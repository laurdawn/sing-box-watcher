package api

import (
	"encoding/json"
	"net/http"

	internalauth "github.com/laurdawn/sing-box-watcher/internal/auth"
	"github.com/laurdawn/sing-box-watcher/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Username != "admin" {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	s.settingsMu.RLock()
	hash := s.settings.PasswordHash
	s.settingsMu.RUnlock()
	if !store.CheckPassword(hash, body.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token := s.authStore.Create()
	internalauth.SetCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := internalauth.GetSessionToken(r)
	if token != "" {
		s.authStore.Delete(token)
	}
	internalauth.ClearCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"username": "admin"})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.NewPassword == "" {
		http.Error(w, "new password is required", http.StatusBadRequest)
		return
	}
	s.settingsMu.RLock()
	hash := s.settings.PasswordHash
	s.settingsMu.RUnlock()
	if !store.CheckPassword(hash, body.OldPassword) {
		http.Error(w, "old password is incorrect", http.StatusUnauthorized)
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	s.settingsMu.Lock()
	s.settings.PasswordHash = string(newHash)
	settings := *s.settings
	s.settingsMu.Unlock()
	if err := store.SaveSettings(s.db, &settings); err != nil {
		http.Error(w, "save settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRegenerateMCPToken(w http.ResponseWriter, r *http.Request) {
	newToken := store.NewMCPToken()
	s.settingsMu.Lock()
	s.settings.MCPToken = newToken
	settings := *s.settings
	s.settingsMu.Unlock()
	if err := store.SaveSettings(s.db, &settings); err != nil {
		http.Error(w, "save settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"mcp_token": newToken})
}
