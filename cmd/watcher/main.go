package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/laurdawn/sing-box-watcher/internal/api"
	"github.com/laurdawn/sing-box-watcher/internal/collector"
	"github.com/laurdawn/sing-box-watcher/internal/config"
	"github.com/laurdawn/sing-box-watcher/internal/geo"
	"github.com/laurdawn/sing-box-watcher/internal/retention"
	"github.com/laurdawn/sing-box-watcher/internal/store"
	"github.com/laurdawn/sing-box-watcher/internal/webfs"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	dbPath := filepath.Join(cfg.DataDir, "watcher.db")
	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	settings, err := store.LoadSettings(db, cfg.DataDir)
	if err != nil {
		log.Fatalf("load settings: %v", err)
	}

	geoDBPath := settings.GeoDBPath
	if err := geo.EnsureDB(geoDBPath, settings.GeoDBURL); err != nil {
		log.Printf("geo db download failed: %v, IP lookup disabled", err)
	} else if err := geo.Init(geoDBPath); err != nil {
		log.Printf("geo db load failed: %v, IP lookup disabled", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	manager := collector.NewManager(settings.Instances, db)
	go manager.Run(ctx)
	go retention.Run(ctx, db, settings.RetentionDuration())

	srv := api.NewServer(cfg, db, settings, manager)

	var staticFS http.FileSystem
	if fs := webfs.FS(); fs != nil {
		staticFS = fs
	}

	httpSrv := &http.Server{
		Addr:    cfg.Listen,
		Handler: srv.Handler(staticFS),
	}

	go func() {
		<-ctx.Done()
		httpSrv.Shutdown(context.Background())
	}()

	log.Printf("sing-box-watcher listening on %s", cfg.Listen)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
