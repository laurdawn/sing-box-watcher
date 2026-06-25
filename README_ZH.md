# sing-box-watcher

为 [sing-box](https://github.com/SagerNet/sing-box) 设计的轻量级监控面板，通过 gRPC 采集流量、连接和代理分组数据，持久化存储到 SQLite。

## 功能

- 实时流量监控（WebSocket 推送）
- 连接历史记录，支持全文搜索和多维度过滤
- 来源 IP 地理位置分析（基于 GeoLite2）
- 代理分组管理：查看延迟、手动选择节点、触发 URL 测速
- 多实例支持：统一管理多台 sing-box 节点
- 配置热重载：保存设置后无需重启
- 前端内嵌：单二进制文件部署
- **MCP Server**：通过标准 MCP 协议让 AI 直接分析流量和连接数据

## 环境要求

- Go 1.21+
- Node.js 18+（仅前端开发时需要）
- sing-box 已启用 `experimental.clash_api` 或 gRPC API

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/laurdawn/sing-box-watcher.git
cd sing-box-watcher
```

### 2. 构建前端

```bash
cd web && npm install && npm run build && cd ..
```

### 3. 构建后端

```bash
go build -o watcher ./cmd/watcher
```

### 4. 配置

```bash
cp config.example.yaml config.yaml
# 按需修改 config.yaml
```

### 5. 运行

```bash
./watcher -config config.yaml
```

浏览器访问 `http://localhost:8080`。

## 配置说明

`config.yaml` 仅控制启动参数，其余配置（实例列表、数据保留天数、GeoIP 路径）在 Web 界面的设置页管理，自动持久化到 SQLite。

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `listen` | `:8080` | HTTP 监听地址 |
| `data_dir` | `./data` | 数据库和 GeoIP 文件存放目录 |

## sing-box API 配置

在 sing-box 配置中启用 gRPC API：

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

然后在 watcher 设置页添加实例，填入 API 地址（如 `http://your-host:9090`）和 secret。

## AI / MCP 集成

sing-box-watcher 内置 MCP Server，可让支持 MCP 协议的 AI（如 Claude）直接查询和分析数据。

### 启用方式

在 Web UI 设置页找到 **AI / MCP** 部分，打开开关即可。MCP Server 地址为：

```
http://your-server:8080/mcp
```

### Claude 配置示例

在 Claude Desktop 或 Claude Code 的 MCP 配置中添加：

```json
{
  "mcpServers": {
    "sing-box": {
      "url": "http://your-server:8080/mcp"
    }
  }
}
```

### 可用工具

| 工具 | 说明 |
|------|------|
| `list_instances` | 列出所有实例及在线状态 |
| `get_service_info` | 查看版本和运行时长 |
| `query_traffic` | 查询历史流量数据 |
| `query_connections` | 分页查询连接记录 |
| `get_active_connections` | 获取当前活跃连接 |
| `get_top_domains` | 访问最多的域名排行 |
| `get_top_outbounds` | 出站代理流量排行 |
| `get_source_regions` | 来源 IP 地域分布 |
| `get_top_source_ips` | 来源 IP 排行 |
| `list_proxy_groups` | 代理分组及延迟 |
| `select_outbound` | 切换代理节点 |
| `lookup_geo` | IP 地理位置查询 |

## Docker 部署

```bash
docker compose up -d
```

详见 `docker-compose.yml`。

## 开发模式

```bash
# 后端
go run ./cmd/watcher -config config.yaml

# 前端开发服务器（自动代理 /api 和 /ws 到 :8080）
cd web && npm run dev
```

前端访问地址：`http://localhost:5173`

## 数据文件

| 路径 | 说明 |
|------|------|
| `data/watcher.db` | SQLite 数据库（流量、连接、设置） |
| `data/GeoLite2-City.mmdb` | GeoIP 数据库（首次启动自动下载） |

## License

MIT


## 环境要求

- Go 1.21+
- Node.js 18+（仅前端开发时需要）
- sing-box 已启用 `experimental.clash_api` 或 gRPC API

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/laurdawn/sing-box-watcher.git
cd sing-box-watcher
```

### 2. 构建前端

```bash
cd web && npm install && npm run build && cd ..
```

### 3. 构建后端

```bash
go build -o watcher ./cmd/watcher
```

### 4. 配置

```bash
cp config.example.yaml config.yaml
# 按需修改 config.yaml
```

### 5. 运行

```bash
./watcher -config config.yaml
```

浏览器访问 `http://localhost:8080`。

## 配置说明

`config.yaml` 仅控制启动参数，其余配置（实例列表、数据保留天数、GeoIP 路径）在 Web 界面的设置页管理，自动持久化到 SQLite。

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `listen` | `:8080` | HTTP 监听地址 |
| `data_dir` | `./data` | 数据库和 GeoIP 文件存放目录 |

## sing-box API 配置

在 sing-box 配置中启用 gRPC API：

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

然后在 watcher 设置页添加实例，填入 API 地址（如 `http://your-host:9090`）和 secret。

## Docker 部署

```bash
docker compose up -d
```

详见 `docker-compose.yml`。

## 开发模式

```bash
# 后端
go run ./cmd/watcher -config config.yaml

# 前端开发服务器（自动代理 /api 和 /ws 到 :8080）
cd web && npm run dev
```

前端访问地址：`http://localhost:5173`

## 数据文件

| 路径 | 说明 |
|------|------|
| `data/watcher.db` | SQLite 数据库（流量、连接、设置） |
| `data/GeoLite2-City.mmdb` | GeoIP 数据库（首次启动自动下载） |

## License

MIT
