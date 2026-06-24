package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/laurdawn/sing-box-watcher/internal/daemon"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type logMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Reset   bool   `json:"reset,omitempty"`
}

func (s *Server) handleWsLog(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 用 WebSocket 请求的 context，断开时 gRPC 流自动取消
	ctx := r.Context()

	s.manager.WithGRPC(instance, func(c daemon.StartedServiceClient, secret string) error {
		authCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+secret)

		stream, err := c.SubscribeLog(authCtx, &emptypb.Empty{})
		if err != nil {
			return err
		}

		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF || ctx.Err() != nil {
					return nil
				}
				return err
			}

			if msg.Reset_ {
				// 通知前端清空，然后补发所有历史日志
				reset := logMessage{Reset: true}
				data, _ := json.Marshal(reset)
				conn.WriteMessage(websocket.TextMessage, data)
			}

			for _, m := range msg.Messages {
				entry := logMessage{
					Level:   m.Level.String(),
					Message: m.Message,
				}
				data, _ := json.Marshal(entry)
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return nil
				}
			}
		}
	})
}
