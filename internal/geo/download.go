package geo

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const defaultDBURL = "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"

// EnsureDB 检查 dbPath 是否存在，不存在则下载。
func EnsureDB(dbPath, dlURL string) error {
	if _, err := os.Stat(dbPath); err == nil {
		return nil
	}

	if dlURL == "" {
		dlURL = defaultDBURL
	}

	log.Printf("geo db not found at %s, downloading from %s ...", dbPath, dlURL)

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	resp, err := http.Get(dlURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: http %d", resp.StatusCode)
	}

	tmp := dbPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}

	n, err := io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return fmt.Errorf("write: %w", err)
	}

	if err := os.Rename(tmp, dbPath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}

	log.Printf("geo db downloaded: %s (%.1f MB)", dbPath, float64(n)/1024/1024)
	return nil
}
