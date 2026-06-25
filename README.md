# sing-box-watcher

A lightweight monitoring dashboard for [sing-box](https://github.com/SagerNet/sing-box), collecting traffic, connections, and proxy group data via gRPC and persisting it to SQLite.

> 中文文档：[README_ZH.md](README_ZH.md)

## Features

- Real-time traffic monitoring with WebSocket push
- Historical connection records with full-text search and multi-dimensional filtering
- Source IP geolocation analysis (powered by GeoLite2)
- Proxy group management — view delays, select outbounds, trigger URL tests
- Multi-instance support — manage multiple sing-box nodes from one dashboard
- Hot-reload configuration — no restart required after saving settings
- Login authentication with session cookie and MCP Bearer Token support
- Log persistence to SQLite with level filtering and keyword search
- Embedded frontend — single binary deployment
- **MCP Server** — let AI (e.g. Claude) query and analyze traffic data directly via the MCP protocol

## Requirements

- Go 1.21+
- Node.js 18+ (for frontend development only)
- sing-box with `experimental.clash_api` enabled

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

### 4. Run

```bash
./watcher
```

Open `http://localhost:8080` in your browser. Default credentials: `admin / admin`.

## Docker

```bash
docker compose up -d
```

See `docker-compose.yml` for configuration options.

## AI / MCP Integration

sing-box-watcher ships with a built-in MCP Server, allowing MCP-compatible AI assistants (e.g. Claude) to query and analyze your proxy data directly.

### Enable

Go to the **AI / MCP** section in the Settings page and toggle the switch on. Copy the generated Bearer Token. The MCP endpoint is:

```
http://your-server:8080/mcp
```

### Claude Configuration

```json
{
  "mcpServers": {
    "sing-box": {
      "type": "http",
      "url": "http://your-server:8080/mcp",
      "headers": {
        "Authorization": "Bearer <your-mcp-token>"
      }
    }
  }
}
```

### Available Tools

| Tool | Description |
|------|-------------|
| `list_instances` | List all instances with online status and current traffic |
| `get_service_info` | Get sing-box version and uptime |
| `query_traffic` | Query historical traffic data by time range |
| `query_connections` | Paginated connection records with filters |
| `get_active_connections` | Currently open connections |
| `get_recent_logs` | Query logs with level filter and keyword search |
| `get_top_domains` | Most accessed domains ranking |
| `get_top_outbounds` | Outbound proxy traffic ranking |
| `get_source_regions` | Source IP geographic distribution |
| `get_top_source_ips` | Top source IPs ranking |
| `list_proxy_groups` | Proxy groups with latency info |
| `select_outbound` | Switch active outbound for a proxy group |
| `lookup_geo` | IP geolocation lookup |

## Development

```bash
# Backend
go run ./cmd/watcher

# Frontend dev server (proxies /api and /ws to :8080)
cd web && npm run dev
```

Frontend is available at `http://localhost:5173`.

## Data

| Path | Description |
|------|-------------|
| `data/watcher.db` | SQLite database (traffic, connections, logs, settings) |
| `data/GeoLite2-City.mmdb` | GeoIP database (auto-downloaded on first run) |

## License

MIT
