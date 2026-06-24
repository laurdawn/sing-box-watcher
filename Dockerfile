# Stage 1: 构建前端（原生 amd64，不需要 QEMU）
FROM --platform=linux/amd64 node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: 交叉编译 Go（原生 amd64 跑编译，输出目标平台二进制）
FROM --platform=linux/amd64 golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./internal/webfs/dist
# TARGETOS/TARGETARCH 由 docker buildx 自动注入
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -tags=prod -ldflags="-s -w" -o watcher ./cmd/watcher

# Stage 3: 最终镜像（目标平台）
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/watcher .
VOLUME ["/app/data"]
EXPOSE 8080
CMD ["./watcher", "-config", "/app/config.yaml"]
