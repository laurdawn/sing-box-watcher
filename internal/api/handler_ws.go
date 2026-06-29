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

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	type snapshot struct{ up, down int64 }
	last := map[string]snapshot{}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}

		stats := s.manager.Stats()
		changed := len(stats) != len(last)
		if !changed {
			for _, st := range stats {
				prev, ok := last[st.Name]
				if !ok || st.CurrentUp != prev.up || st.CurrentDown != prev.down {
					changed = true
					break
				}
			}
		}
		if !changed {
			continue
		}

		now := time.Now().Unix()
		payload := make([]trafficPush, 0, len(stats))
		for _, st := range stats {
			payload = append(payload, trafficPush{
				Instance: st.Name,
				Up:       st.CurrentUp,
				Down:     st.CurrentDown,
				TS:       now,
			})
			last[st.Name] = snapshot{up: st.CurrentUp, down: st.CurrentDown}
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
