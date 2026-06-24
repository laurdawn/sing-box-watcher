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

type ServiceStatus struct {
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	Online    bool   `json:"online"`
	UpdatedAt int64  `json:"updated_at"`
}

type StatusCollector struct {
	instance string
	apiURL   string
	secret   string

	mu     sync.RWMutex
	status ServiceStatus
}

func NewStatusCollector(instance, apiURL, secret string) *StatusCollector {
	return &StatusCollector{
		instance: instance,
		apiURL:   apiURL,
		secret:   secret,
		status:   ServiceStatus{Status: "UNKNOWN"},
	}
}

func (c *StatusCollector) Current() ServiceStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *StatusCollector) setOffline() {
	c.mu.Lock()
	c.status = ServiceStatus{Status: "OFFLINE", Online: false, UpdatedAt: time.Now().UnixMilli()}
	c.mu.Unlock()
}

func (c *StatusCollector) Run(ctx context.Context) {
	for {
		if err := c.connect(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[%s] status collector error: %v, retrying...", c.instance, err)
		}
		c.setOffline()
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

func (c *StatusCollector) connect(ctx context.Context) error {
	conn, err := newGRPCConn(c.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := daemon.NewStartedServiceClient(conn)
	authCtx := withAuth(ctx, c.secret)

	stream, err := client.SubscribeServiceStatus(authCtx, &emptypb.Empty{})
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
		c.mu.Lock()
		c.status = ServiceStatus{
			Status:    msg.Status.String(),
			Error:     msg.ErrorMessage,
			Online:    msg.Status == daemon.ServiceStatus_STARTED,
			UpdatedAt: time.Now().UnixMilli(),
		}
		c.mu.Unlock()
	}
}
