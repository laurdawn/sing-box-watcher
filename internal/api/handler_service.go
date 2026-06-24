package api

import (
	"context"
	"net/http"
	"time"

	"github.com/sagernet/sing-box/daemon"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serviceInfoResponse struct {
	Version            string   `json:"version"`
	APIVersion         int32    `json:"api_version"`
	StartedAt          int64    `json:"started_at_ms"`
	UptimeSeconds      int64    `json:"uptime_seconds"`
	Status             string   `json:"status"`
	Online             bool     `json:"online"`
	DeprecatedWarnings []string `json:"deprecated_warnings"`
}

func (s *Server) handleServiceInfo(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")

	info := serviceInfoResponse{}

	// 从 status collector 取实时状态
	stats := s.manager.Stats()
	for _, st := range stats {
		if st.Name == instance {
			info.Status = st.Status
			info.Online = st.Online
			break
		}
	}

	// Unary 调用获取版本、启动时间、废弃警告
	_ = s.manager.WithClient(instance, func(c daemon.StartedServiceClient, ctx context.Context) error {
		if v, err := c.GetVersion(ctx, &emptypb.Empty{}); err == nil {
			info.Version = v.Version
			info.APIVersion = v.ApiVersion
		}
		if sa, err := c.GetStartedAt(ctx, &emptypb.Empty{}); err == nil {
			info.StartedAt = sa.StartedAt
			info.UptimeSeconds = int64(time.Since(time.UnixMilli(sa.StartedAt)).Seconds())
		}
		if dw, err := c.GetDeprecatedWarnings(ctx, &emptypb.Empty{}); err == nil {
			for _, w := range dw.Warnings {
				info.DeprecatedWarnings = append(info.DeprecatedWarnings, w.Message)
			}
		}
		return nil
	})

	if info.DeprecatedWarnings == nil {
		info.DeprecatedWarnings = []string{}
	}

	writeJSON(w, http.StatusOK, info)
}
