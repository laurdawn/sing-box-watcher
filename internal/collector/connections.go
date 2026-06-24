package collector

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/sagernet/sing-box/daemon"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type ConnectionCollector struct {
	instance string
	apiURL   string
	secret   string
	db       *sql.DB

	mu     sync.RWMutex
	active map[string]*store.Connection
}

func NewConnectionCollector(instance, apiURL, secret string, db *sql.DB) *ConnectionCollector {
	return &ConnectionCollector{
		instance: instance,
		apiURL:   apiURL,
		secret:   secret,
		db:       db,
		active:   make(map[string]*store.Connection),
	}
}

func (c *ConnectionCollector) ActiveCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.active)
}

func (c *ConnectionCollector) Run(ctx context.Context) {
	for {
		if err := c.connect(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[%s] connections collector error: %v, retrying...", c.instance, err)
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

func (c *ConnectionCollector) connect(ctx context.Context) error {
	conn, err := newGRPCConn(c.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := daemon.NewStartedServiceClient(conn)
	authCtx := withAuth(ctx, c.secret)

	stream, err := client.SubscribeConnections(authCtx, &daemon.SubscribeConnectionsRequest{Interval: int64(time.Second)})
	if err != nil {
		return err
	}

	log.Printf("[%s] connections collector connected", c.instance)

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				return nil
			}
			return err
		}
		c.processEvents(msg)
	}
}

func (c *ConnectionCollector) processEvents(msg *daemon.ConnectionEvents) {
	if msg.Reset_ {
		// 服务端重置，关闭所有已追踪连接
		c.mu.Lock()
		now := time.Now().UnixMilli()
		for id, conn := range c.active {
			store.CloseConnection(c.db, id, now, conn.Upload, conn.Download)
		}
		c.active = make(map[string]*store.Connection)
		c.mu.Unlock()
	}

	for _, event := range msg.Events {
		switch event.Type {
		case daemon.ConnectionEventType_CONNECTION_EVENT_NEW:
			conn := protoToStore(c.instance, event.Connection)
			if conn == nil {
				continue
			}
			c.mu.Lock()
			c.active[conn.ID] = conn
			c.mu.Unlock()
			if err := store.UpsertConnection(c.db, conn); err != nil {
				log.Printf("[%s] upsert connection error: %v", c.instance, err)
			}


		case daemon.ConnectionEventType_CONNECTION_EVENT_UPDATE:
			c.mu.Lock()
			if conn, ok := c.active[event.Id]; ok {
				conn.Upload += event.UplinkDelta
				conn.Download += event.DownlinkDelta
				store.UpsertConnection(c.db, conn)
			}
			c.mu.Unlock()

		case daemon.ConnectionEventType_CONNECTION_EVENT_CLOSED:
			c.mu.Lock()
			conn, ok := c.active[event.Id]
			if ok {
				delete(c.active, event.Id)
			}
			c.mu.Unlock()
			if ok {
				closedAt := event.ClosedAt // proto 已经是毫秒
				if closedAt == 0 {
					closedAt = time.Now().UnixMilli()
				}
				store.CloseConnection(c.db, event.Id, closedAt, conn.Upload, conn.Download)
			}
		}
	}
}

func protoToStore(instance string, p *daemon.Connection) *store.Connection {
	if p == nil {
		return nil
	}
	chains, _ := json.Marshal(p.ChainList)
	processPath := ""
	if p.ProcessInfo != nil {
		processPath = p.ProcessInfo.ProcessPath
		if processPath == "" && len(p.ProcessInfo.PackageNames) > 0 {
			processPath = p.ProcessInfo.PackageNames[0]
		}
	}
	// destination 格式为 "ip:port" 或 "domain:port"
	destIP, destPort := splitAddr(p.Destination)
	srcIP, srcPort := splitAddr(p.Source)

	host := p.Domain
	if host == "" {
		host = destIP
	}

	return &store.Connection{
		ID:           p.Id,
		Instance:     instance,
		Network:      p.Network,
		Inbound:      p.Inbound,
		InboundType:  p.InboundType,
		Outbound:     p.Outbound,
		OutboundType: p.OutboundType,
		SourceIP:     srcIP,
		SourcePort:   srcPort,
		DestIP:       destIP,
		DestPort:     destPort,
		Host:         host,
		ProcessPath:  processPath,
		Rule:         p.Rule,
		Chains:       string(chains),
		Upload:    p.Uplink,
		Download:  p.Downlink,
		StartedAt: p.CreatedAt, // 毫秒
	}
}

func splitAddr(addr string) (ip string, port int) {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			ip = addr[:i]
			for _, ch := range addr[i+1:] {
				if ch >= '0' && ch <= '9' {
					port = port*10 + int(ch-'0')
				}
			}
			return
		}
	}
	return addr, 0
}
