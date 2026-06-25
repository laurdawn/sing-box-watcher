package collector

import (
	"context"
	"database/sql"
	"sort"
	"sync"

	"github.com/laurdawn/sing-box-watcher/internal/daemon"
	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type InstanceStats struct {
	Name        string `json:"name"`
	CurrentUp   int64  `json:"current_up"`
	CurrentDown int64  `json:"current_down"`
	ActiveConns int    `json:"active_connections"`
	Online      bool   `json:"online"`
	Status      string `json:"status"`
}

type instanceEntry struct {
	traffic     *TrafficCollector
	connections *ConnectionCollector
	status      *StatusCollector
	groups      *GroupsCollector
	logs        *LogCollector
	cancel      context.CancelFunc
}

type Manager struct {
	db      *sql.DB
	rootCtx context.Context

	mu        sync.RWMutex
	instances map[string]*instanceEntry
}

func NewManager(instances []store.Instance, db *sql.DB) *Manager {
	m := &Manager{
		db:        db,
		instances: make(map[string]*instanceEntry),
	}
	for _, inst := range instances {
		m.instances[inst.Name] = newEntry(inst, db)
	}
	return m
}

func newEntry(inst store.Instance, db *sql.DB) *instanceEntry {
	return &instanceEntry{
		traffic:     NewTrafficCollector(inst.Name, inst.API, inst.Secret, db),
		connections: NewConnectionCollector(inst.Name, inst.API, inst.Secret, db),
		status:      NewStatusCollector(inst.Name, inst.API, inst.Secret),
		groups:      NewGroupsCollector(inst.Name, inst.API, inst.Secret),
		logs:        NewLogCollector(inst.Name, inst.API, inst.Secret),
	}
}

func (m *Manager) Run(ctx context.Context) {
	m.rootCtx = ctx
	m.mu.RLock()
	var entries []*instanceEntry
	for _, e := range m.instances {
		entries = append(entries, e)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, e := range entries {
		m.startEntry(ctx, e, &wg)
	}
	wg.Wait()
}

func (m *Manager) startEntry(rootCtx context.Context, e *instanceEntry, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(rootCtx)
	e.cancel = cancel

	wg.Add(5)
	go func() { defer wg.Done(); e.traffic.Run(ctx) }()
	go func() { defer wg.Done(); e.connections.Run(ctx) }()
	go func() { defer wg.Done(); e.status.Run(ctx) }()
	go func() { defer wg.Done(); e.groups.Run(ctx) }()
	go func() { defer wg.Done(); e.logs.Run(ctx) }()
}

func (m *Manager) Reload(newInstances []store.Instance) {
	if m.rootCtx == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	newMap := make(map[string]store.Instance, len(newInstances))
	for _, inst := range newInstances {
		newMap[inst.Name] = inst
	}

	for name, e := range m.instances {
		inst, exists := newMap[name]
		if !exists || inst.API != e.traffic.apiURL || inst.Secret != e.traffic.secret {
			if e.cancel != nil {
				e.cancel()
			}
			delete(m.instances, name)
		}
	}

	var wg sync.WaitGroup
	for _, inst := range newInstances {
		if _, exists := m.instances[inst.Name]; !exists {
			e := newEntry(inst, m.db)
			m.instances[inst.Name] = e
			m.startEntry(m.rootCtx, e, &wg)
		}
	}
}

func (m *Manager) Stats() []InstanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]InstanceStats, 0, len(m.instances))
	for name, e := range m.instances {
		up, down := e.traffic.Current()
		svc := e.status.Current()
		stats = append(stats, InstanceStats{
			Name:        name,
			CurrentUp:   up,
			CurrentDown: down,
			ActiveConns: e.connections.ActiveCount(),
			Online:      svc.Online,
			Status:      svc.Status,
		})
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].Name < stats[j].Name })
	return stats
}

func (m *Manager) CurrentTraffic(instance string) (up, down int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if e, ok := m.instances[instance]; ok {
		return e.traffic.Current()
	}
	return 0, 0
}

func (m *Manager) Groups(instance string) *GroupsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if e, ok := m.instances[instance]; ok {
		s := e.groups.Snapshot()
		return &s
	}
	return nil
}

func (m *Manager) GroupsNotify(instance string) <-chan struct{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if e, ok := m.instances[instance]; ok {
		return e.groups.Subscribe()
	}
	return nil
}

func (m *Manager) WithClient(instance string, fn func(daemon.StartedServiceClient, context.Context) error) error {
	m.mu.RLock()
	e, ok := m.instances[instance]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	conn, err := newGRPCConn(e.traffic.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := daemon.NewStartedServiceClient(conn)
	ctx := withAuth(m.rootCtx, e.traffic.secret)
	return fn(client, ctx)
}

func (m *Manager) WithGRPC(instance string, fn func(daemon.StartedServiceClient, string) error) error {
	m.mu.RLock()
	e, ok := m.instances[instance]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	conn, err := newGRPCConn(e.traffic.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := daemon.NewStartedServiceClient(conn)
	return fn(client, e.traffic.secret)
}

func (m *Manager) RecentLogs(instance string, n int, level, keyword string) []LogEntry {
	m.mu.RLock()
	e, ok := m.instances[instance]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	return e.logs.Recent(n, level, keyword)
}
