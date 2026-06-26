# sing-box-watcher

为 [sing-box](https://github.com/SagerNet/sing-box) 设计的轻量级监控面板，采集流量、连接和代理分组数据，持久化到 SQLite。

## 功能

- 实时流量监控（WebSocket 推送）
- 连接历史记录，支持全文搜索和多维度过滤
- 来源 IP 地理位置分析（基于 GeoLite2）
- 代理分组管理：查看延迟、选择节点、触发 URL 测速
- 多实例支持，登录认证，配置热重载
- 日志持久化，支持按级别过滤和关键字搜索
- 前端内嵌，单二进制部署
- **MCP Server**：让 AI（如 Claude）直接查询和分析流量数据

## 快速开始

```bash
git clone https://github.com/laurdawn/sing-box-watcher.git
cd sing-box-watcher
cd web && npm install && npm run build && cd ..
go build -o watcher ./cmd/watcher
./watcher
```

浏览器访问 `http://localhost:8080`，默认账号：`admin / admin`。

## Docker

```bash
docker compose up -d
```

## AI / MCP 集成

在设置页 **AI / MCP** 部分打开开关，复制 Bearer Token。MCP 地址：

```
http://your-server:8080/mcp
```

Claude 配置示例：

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

可用工具：`list_instances` · `get_service_info` · `query_traffic` · `query_connections` · `get_active_connections` · `get_recent_logs` · `get_top_domains` · `get_top_outbounds` · `get_source_regions` · `get_top_source_ips` · `list_proxy_groups` · `select_outbound` · `lookup_geo`

## 开发

```bash
go run ./cmd/watcher          # 后端
cd web && npm run dev         # 前端 → http://localhost:5173
```

## 数据文件

| 路径 | 说明 |
|------|------|
| `data/watcher.db` | SQLite 数据库 |
| `data/GeoLite2-City.mmdb` | GeoIP 数据库（首次启动自动下载） |

## License

MIT

首发于 [linux.do](https://linux.do) 论坛
