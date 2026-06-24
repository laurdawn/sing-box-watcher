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
	defer ticker.Stop()

	clean := func() {
		before := time.Now().Add(-retention).UnixMilli()
		if err := store.DeleteOldTraffic(db, before); err != nil {
			log.Printf("retention: delete old traffic: %v", err)
		}
		if err := store.DeleteOldConnections(db, before); err != nil {
			log.Printf("retention: delete old connections: %v", err)
		}
	}

	clean()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			clean()
		}
	}
}
