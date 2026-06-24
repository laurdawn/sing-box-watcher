# Stage 1: 构建前端
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: 构建 Go（带嵌入前端）
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 将前端产物复制到 embed 目录
COPY --from=frontend /app/web/dist ./internal/webfs/dist
RUN CGO_ENABLED=0 go build -tags=prod -ldflags="-s -w" -o watcher ./cmd/watcher

# Stage 3: 最终镜像 (~20MB)
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/watcher .
VOLUME ["/app/data"]
EXPOSE 8080
CMD ["./watcher", "-config", "/app/config.yaml"]
