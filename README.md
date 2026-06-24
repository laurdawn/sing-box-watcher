# sing-box-watcher

A lightweight monitoring dashboard for [sing-box](https://github.com/SagerNet/sing-box), collecting traffic, connections, and proxy group data via gRPC and persisting it to SQLite.

> 中文文档：[README_ZH.md](README_ZH.md)

## Features

- Real-time traffic monitoring with WebSocket push
- Historical connection records with full-text search and filtering
- Source IP geolocation analysis (powered by GeoLite2)
- Proxy group management — view delays, select outbounds, trigger URL tests
- Multi-instance support — manage multiple sing-box nodes from one dashboard
- Hot-reload configuration — no restart required after saving settings
- Embedded frontend — single binary deployment

## Screenshots

> Dashboard / Proxies / Analysis / Connections / Settings

## Requirements

- Go 1.21+
- Node.js 18+ (for frontend development only)
- sing-box with `experimental.clash_api` or gRPC API enabled

## Quick Start

### 1. Clone

```bash
git clone https://github.com/laurdawn/sing-box-watcher.git
cd sing-box-watcher
```

### 2. Build frontend

```bash
cd web && npm install && npm run build && cd ..
```

### 3. Build backend

```bash
go build -o watcher ./cmd/watcher
```

### 4. Configure

```bash
cp config.example.yaml config.yaml
# Edit config.yaml as needed
```

### 5. Run

```bash
./watcher -config config.yaml
```

Open `http://localhost:8080` in your browser.

## Configuration

`config.yaml` controls startup parameters only. All other settings (instances, retention days, GeoIP path) are managed from the web UI and persisted to SQLite.

| Field | Default | Description |
|-------|---------|-------------|
| `listen` | `:8080` | HTTP listen address |
| `data_dir` | `./data` | Directory for database and GeoIP files |

## sing-box API Setup

Enable the gRPC API in your sing-box config:

```json
{
  "experimental": {
    "clash_api": {
      "external_controller": "0.0.0.0:9090",
      "secret": "your-secret"
    }
  }
}
```

Then add the instance in the watcher Settings page: fill in the API address (e.g. `http://your-host:9090`) and the secret.

## Docker

```bash
docker compose up -d
```

See `docker-compose.yml` for configuration options.

## Development

```bash
# Backend (with dev file serving)
go run ./cmd/watcher -config config.yaml

# Frontend dev server (proxies /api and /ws to :8080)
cd web && npm run dev
```

Frontend is available at `http://localhost:5173`.

## Data

| Path | Description |
|------|-------------|
| `data/watcher.db` | SQLite database (traffic, connections, settings) |
| `data/GeoLite2-City.mmdb` | GeoIP database (auto-downloaded on first run) |

## License

MIT
