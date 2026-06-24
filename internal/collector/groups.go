package collector

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/daemon"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GroupsSnapshot struct {
	Groups    []*daemon.Group `json:"groups,omitempty"`
	UpdatedAt int64           `json:"updated_at"`
}

type GroupsCollector struct {
	instance string
	apiURL   string
	secret   string

	mu       sync.RWMutex
	snapshot GroupsSnapshot

	// 变更通知，写入不阻塞
	notify chan struct{}
}

func NewGroupsCollector(instance, apiURL, secret string) *GroupsCollector {
	return &GroupsCollector{
		instance: instance,
		apiURL:   apiURL,
		secret:   secret,
		notify:   make(chan struct{}, 1),
	}
}

func (c *GroupsCollector) Snapshot() GroupsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot
}

// Subscribe 返回一个 channel，每次分组变更时收到信号。
func (c *GroupsCollector) Subscribe() <-chan struct{} {
	return c.notify
}

func (c *GroupsCollector) Run(ctx context.Context) {
	for {
		if err := c.connect(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[%s] groups collector error: %v, retrying...", c.instance, err)
		}
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnectDelay()):
		}
	}
}

func (c *GroupsCollector) connect(ctx context.Context) error {
	conn, err := newGRPCConn(c.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := daemon.NewStartedServiceClient(conn)
	authCtx := withAuth(ctx, c.secret)

	stream, err := client.SubscribeGroups(authCtx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	log.Printf("[%s] groups collector connected", c.instance)

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}
		c.mu.Lock()
		c.snapshot = GroupsSnapshot{
			Groups:    msg.Group,
			UpdatedAt: time.Now().UnixMilli(),
		}
		c.mu.Unlock()

		// 非阻塞通知
		select {
		case c.notify <- struct{}{}:
		default:
		}
	}
}
