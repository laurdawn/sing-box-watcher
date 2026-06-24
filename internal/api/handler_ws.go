package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/laurdawn/sing-box-watcher/internal/collector"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type trafficPush struct {
	Instance string `json:"instance"`
	Up       int64  `json:"up"`
	Down     int64  `json:"down"`
	TS       int64  `json:"ts"`
}

func (s *Server) handleWsTraffic(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}

		stats := s.manager.Stats()
		payload := make([]trafficPush, 0, len(stats))
		for _, st := range stats {
			payload = append(payload, trafficPush{
				Instance: st.Name,
				Up:       st.CurrentUp,
				Down:     st.CurrentDown,
				TS:       time.Now().Unix(),
			})
		}
		data, _ := json.Marshal(payload)
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

// instanceStats 向前端返回所有实例统计（含活跃连接数）
func instanceStatsFromManager(manager *collector.Manager) []collector.InstanceStats {
	return manager.Stats()
}
