package collector

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/daemon"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type activeConn struct {
	Upload    int64
	Download  int64
	StartedAt int64
}

type ConnectionCollector struct {
	instance string
	apiURL   string
	secret   string
	db       *sql.DB

	mu     sync.RWMutex
	active map[string]*activeConn
	dirty  map[string]struct{}
}

func NewConnectionCollector(instance, apiURL, secret string, db *sql.DB) *ConnectionCollector {
	return &ConnectionCollector{
		instance: instance,
		apiURL:   apiURL,
		secret:   secret,
		db:       db,
		active:   make(map[string]*activeConn),
		dirty:    make(map[string]struct{}),
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

	// flush dirty connections to DB every 2 seconds
	stopFlush := make(chan struct{})
	defer close(stopFlush)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.flushDirty()
			case <-stopFlush:
				return
			}
		}
	}()

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

func (c *ConnectionCollector) flushDirty() {
	c.mu.Lock()
	if len(c.dirty) == 0 {
		c.mu.Unlock()
		return
	}
	snapshot := make(map[string]activeConn, len(c.dirty))
	for id := range c.dirty {
		if ac, ok := c.active[id]; ok {
			snapshot[id] = *ac
		}
	}
	c.dirty = make(map[string]struct{})
	c.mu.Unlock()

	for id, ac := range snapshot {
		if err := store.UpdateConnectionTraffic(c.db, id, ac.Upload, ac.Download); err != nil {
			log.Printf("[%s] update connection traffic: %v", c.instance, err)
		}
	}
}

func (c *ConnectionCollector) processEvents(msg *daemon.ConnectionEvents) {
	if msg.Reset_ {
		c.mu.Lock()
		now := time.Now().UnixMilli()
		for id, ac := range c.active {
			store.CloseConnection(c.db, id, now, ac.Upload, ac.Download)
		}
		c.active = make(map[string]*activeConn)
		c.dirty = make(map[string]struct{})
		c.mu.Unlock()
	}

	for _, event := range msg.Events {
		switch event.Type {
		case daemon.ConnectionEventType_CONNECTION_EVENT_NEW:
			conn := protoToStore(c.instance, event.Connection)
			if conn == nil {
				continue
			}
			ac := &activeConn{
				Upload:    conn.Upload,
				Download:  conn.Download,
				StartedAt: conn.StartedAt,
			}
			c.mu.Lock()
			c.active[conn.ID] = ac
			c.mu.Unlock()
			if err := store.UpsertConnection(c.db, conn); err != nil {
				log.Printf("[%s] upsert connection error: %v", c.instance, err)
			}

		case daemon.ConnectionEventType_CONNECTION_EVENT_UPDATE:
			c.mu.Lock()
			if ac, ok := c.active[event.Id]; ok {
				ac.Upload += event.UplinkDelta
				ac.Download += event.DownlinkDelta
				c.dirty[event.Id] = struct{}{}
			}
			c.mu.Unlock()

		case daemon.ConnectionEventType_CONNECTION_EVENT_CLOSED:
			c.mu.Lock()
			ac, ok := c.active[event.Id]
			if ok {
				delete(c.active, event.Id)
				delete(c.dirty, event.Id)
			}
			c.mu.Unlock()
			if ok {
				closedAt := event.ClosedAt
				if closedAt == 0 {
					closedAt = time.Now().UnixMilli()
				}
				store.CloseConnection(c.db, event.Id, closedAt, ac.Upload, ac.Download)
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
		StartedAt: p.CreatedAt,
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
