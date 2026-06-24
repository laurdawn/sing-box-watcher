package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/laurdawn/sing-box-watcher/internal/daemon"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	snap := s.manager.Groups(instance)
	if snap == nil {
		writeJSON(w, http.StatusOK, map[string]any{"groups": []any{}, "updated_at": 0})
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

type selectOutboundRequest struct {
	Instance   string `json:"instance"`
	GroupTag   string `json:"group_tag"`
	OutboundTag string `json:"outbound_tag"`
}

func (s *Server) handleSelectOutbound(w http.ResponseWriter, r *http.Request) {
	var req selectOutboundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	err := s.manager.WithClient(req.Instance, func(c daemon.StartedServiceClient, ctx context.Context) error {
		_, err := c.SelectOutbound(ctx, &daemon.SelectOutboundRequest{
			GroupTag:    req.GroupTag,
			OutboundTag: req.OutboundTag,
		})
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type urlTestRequest struct {
	Instance    string `json:"instance"`
	OutboundTag string `json:"outbound_tag"`
}

func (s *Server) handleURLTest(w http.ResponseWriter, r *http.Request) {
	var req urlTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	err := s.manager.WithClient(req.Instance, func(c daemon.StartedServiceClient, ctx context.Context) error {
		_, err := c.URLTest(ctx, &daemon.URLTestRequest{OutboundTag: req.OutboundTag})
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleWsGroups(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 立即推送当前快照
	if snap := s.manager.Groups(instance); snap != nil {
		data, _ := json.Marshal(snap)
		conn.WriteMessage(websocket.TextMessage, data)
	}

	notify := s.manager.GroupsNotify(instance)
	if notify == nil {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case _, ok := <-notify:
			if !ok {
				return
			}
			snap := s.manager.Groups(instance)
			if snap == nil {
				continue
			}
			data, _ := json.Marshal(snap)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

// handleGetOutbounds 返回所有出站节点（平铺列表，非分组）
func (s *Server) handleGetOutbounds(w http.ResponseWriter, r *http.Request) {
	instance := r.URL.Query().Get("instance")
	type outboundItem struct {
		Tag          string `json:"tag"`
		Type         string `json:"type"`
		URLTestTime  int64  `json:"url_test_time"`
		URLTestDelay int32  `json:"url_test_delay"`
	}
	var items []outboundItem
	s.manager.WithClient(instance, func(c daemon.StartedServiceClient, ctx context.Context) error {
		stream, err := c.SubscribeOutbounds(ctx, &emptypb.Empty{})
		if err != nil {
			return err
		}
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		for _, o := range msg.Outbounds {
			items = append(items, outboundItem{
				Tag:          o.Tag,
				Type:         o.Type,
				URLTestTime:  o.UrlTestTime,
				URLTestDelay: o.UrlTestDelay,
			})
		}
		return nil
	})
	if items == nil {
		items = []outboundItem{}
	}
	writeJSON(w, http.StatusOK, items)
}
