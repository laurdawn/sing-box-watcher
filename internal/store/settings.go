package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Settings 存储运行时可修改的配置，持久化到 SQLite。
type Settings struct {
	RetentionDays      int        `json:"retention_days"`
	GeoDBPath          string     `json:"geo_db_path"`
	GeoDBURL           string     `json:"geo_db_url"`
	Instances          []Instance `json:"instances"`
	MCPEnabled         bool       `json:"mcp_enabled"`
	LogPersistEnabled  bool       `json:"log_persist_enabled"`
	LogPersistMinLevel string     `json:"log_persist_min_level"` // 默认 "WARN"
}

type Instance struct {
	Name   string `json:"name"   yaml:"name"`
	API    string `json:"api"    yaml:"api"`
	Secret string `json:"secret" yaml:"secret"`
}

func initSettingsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
)`)
	return err
}

func (s *Settings) RetentionDuration() time.Duration {
	return time.Duration(s.RetentionDays) * 24 * time.Hour
}

func LoadSettings(db *sql.DB, dataDir string) (*Settings, error) {
	s := &Settings{
		RetentionDays: 7,
		GeoDBPath:     dataDir + "/GeoLite2-City.mmdb",
		Instances:     []Instance{},
	}
	row := db.QueryRow(`SELECT value FROM settings WHERE key = 'config'`)
	var raw string
	if err := row.Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return s, nil
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(raw), s); err != nil {
		return nil, err
	}
	if s.RetentionDays <= 0 {
		s.RetentionDays = 7
	}
	if s.GeoDBPath == "" {
		s.GeoDBPath = dataDir + "/GeoLite2-City.mmdb"
	}
	if s.Instances == nil {
		s.Instances = []Instance{}
	}
	return s, nil
}

func SaveSettings(db *sql.DB, s *Settings) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO settings(key, value) VALUES('config', ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value`, string(data))
	return err
}
