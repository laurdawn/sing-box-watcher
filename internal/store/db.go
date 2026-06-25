package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	if err := initSettingsTable(db); err != nil {
		return err
	}
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS traffic_stats (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    instance    TEXT NOT NULL,
    ts          INTEGER NOT NULL,
    upload      INTEGER NOT NULL,
    download    INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_traffic_instance_ts ON traffic_stats(instance, ts);

CREATE TABLE IF NOT EXISTS connections (
    id              TEXT PRIMARY KEY,
    instance        TEXT NOT NULL,
    network         TEXT,
    inbound         TEXT,
    inbound_type    TEXT,
    outbound        TEXT,
    outbound_type   TEXT,
    source_ip       TEXT,
    source_port     INTEGER,
    dest_ip         TEXT,
    dest_port       INTEGER,
    host            TEXT,
    process_path    TEXT,
    rule            TEXT,
    chains          TEXT,
    upload          INTEGER DEFAULT 0,
    download        INTEGER DEFAULT 0,
    started_at      INTEGER NOT NULL,
    closed_at       INTEGER
);
CREATE INDEX IF NOT EXISTS idx_conn_instance_started ON connections(instance, started_at);
CREATE INDEX IF NOT EXISTS idx_conn_host ON connections(host);
CREATE INDEX IF NOT EXISTS idx_conn_dest_ip ON connections(dest_ip);
CREATE INDEX IF NOT EXISTS idx_conn_inbound ON connections(instance, inbound_type);
CREATE INDEX IF NOT EXISTS idx_conn_outbound ON connections(instance, outbound);

CREATE TABLE IF NOT EXISTS logs (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    instance  TEXT NOT NULL,
    ts        INTEGER NOT NULL,
    level     TEXT NOT NULL,
    message   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_logs_instance_ts ON logs(instance, ts);
CREATE INDEX IF NOT EXISTS idx_logs_level       ON logs(instance, level, ts);
`)
	return err
}
