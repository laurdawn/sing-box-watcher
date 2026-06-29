package retention

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

func Run(ctx context.Context, db *sql.DB, retention time.Duration) {
	ticker := time.NewTicker(time.Hour)
	vacuumTicker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	defer vacuumTicker.Stop()

	clean := func() {
		before := time.Now().Add(-retention).UnixMilli()
		if err := store.DeleteOldTraffic(db, before); err != nil {
			log.Printf("retention: delete old traffic: %v", err)
		}
		if err := store.DeleteOldConnections(db, before); err != nil {
			log.Printf("retention: delete old connections: %v", err)
		}
		if err := store.DeleteOrphanConnections(db, before); err != nil {
			log.Printf("retention: delete orphan connections: %v", err)
		}
		if err := store.DeleteOldLogs(db, before); err != nil {
			log.Printf("retention: delete old logs: %v", err)
		}
		if _, err := db.Exec("PRAGMA wal_checkpoint(PASSIVE)"); err != nil {
			log.Printf("retention: wal_checkpoint: %v", err)
		}
	}

	clean()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clean()
		case <-vacuumTicker.C:
			if _, err := db.Exec("VACUUM"); err != nil {
				log.Printf("retention: vacuum: %v", err)
			}
			// VACUUM rebuilds the DB file and resets journal_mode to DELETE; restore WAL
			if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
				log.Printf("retention: restore wal mode: %v", err)
			}
		}
	}
}
