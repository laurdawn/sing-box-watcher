package collector

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// newGRPCConn 创建到 sing-box API 的 gRPC 连接，支持 TLS 和明文。
func newGRPCConn(apiURL string) (*grpc.ClientConn, error) {
	host, useTLS := parseAPIURL(apiURL)
	var cred credentials.TransportCredentials
	if useTLS {
		cred = credentials.NewTLS(&tls.Config{})
	} else {
		cred = insecure.NewCredentials()
	}
	return grpc.NewClient(host, grpc.WithTransportCredentials(cred))
}

// parseAPIURL 把 http(s)://host:port 或 host:port 解析成 gRPC target。
func parseAPIURL(apiURL string) (target string, useTLS bool) {
	switch {
	case len(apiURL) > 8 && apiURL[:8] == "https://":
		return apiURL[8:], true
	case len(apiURL) > 7 && apiURL[:7] == "http://":
		return apiURL[7:], false
	default:
		return apiURL, false
	}
}

// withAuth 将 Bearer token 注入到 gRPC 请求 metadata。
func withAuth(ctx context.Context, secret string) context.Context {
	if secret == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+secret)
}
