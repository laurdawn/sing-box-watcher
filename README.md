# sing-box-watcher

A lightweight monitoring dashboard for [sing-box](https://github.com/SagerNet/sing-box), collecting traffic, connections, and proxy group data, persisted to SQLite.

> 中文文档：[README_ZH.md](README_ZH.md)

## Features

- Real-time traffic monitoring via WebSocket
- Connection history with full-text search and filtering
- Source IP geolocation analysis (GeoLite2)
- Proxy group management — latency, node selection, URL tests
- Multi-instance support, login auth, hot-reload config
- Log persistence with level filtering and keyword search
- Embedded frontend, single binary deployment
- **MCP Server** — let AI (e.g. Claude) query traffic data directly

## Quick Start

```bash
git clone https://github.com/laurdawn/sing-box-watcher.git
cd sing-box-watcher
cd web && npm install && npm run build && cd ..
go build -o watcher ./cmd/watcher
./watcher
```

Open `http://localhost:8080`. Default credentials: `admin / admin`.

## Docker

```bash
docker compose up -d
```

## AI / MCP Integration

Enable the MCP Server in Settings → **AI / MCP**, then copy the Bearer Token. Endpoint:

```
http://your-server:8080/mcp
```

Claude config:

```json
{
  "mcpServers": {
    "sing-box": {
      "type": "http",
      "url": "http://your-server:8080/mcp",
      "headers": { "Authorization": "Bearer <your-mcp-token>" }
    }
  }
}
```

Available tools: `list_instances` · `get_service_info` · `query_traffic` · `query_connections` · `get_active_connections` · `get_recent_logs` · `get_top_domains` · `get_top_outbounds` · `get_source_regions` · `get_top_source_ips` · `list_proxy_groups` · `select_outbound` · `lookup_geo`

## Development

```bash
go run ./cmd/watcher          # backend
cd web && npm run dev         # frontend → http://localhost:5173
```

## Data

| Path | Description |
|------|-------------|
| `data/watcher.db` | SQLite database |
| `data/GeoLite2-City.mmdb` | GeoIP database (auto-downloaded on first run) |

## License

MIT
