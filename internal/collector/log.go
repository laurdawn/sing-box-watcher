package collector

import (
	"context"
	"database/sql"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/daemon"
	"github.com/laurdawn/sing-box-watcher/internal/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

const logBufferSize = 500

type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type LogCollector struct {
	instance string
	apiURL   string
	secret   string
	db       *sql.DB

	cfgMu           sync.RWMutex
	persistEnabled  bool
	persistMinLevel string

	mu  sync.RWMutex
	buf []LogEntry
}

func NewLogCollector(instance, apiURL, secret string, db *sql.DB) *LogCollector {
	return &LogCollector{
		instance:        instance,
		apiURL:          apiURL,
		secret:          secret,
		db:              db,
		persistMinLevel: "WARN",
		buf:             make([]LogEntry, 0, logBufferSize),
	}
}

func (c *LogCollector) UpdateConfig(enabled bool, minLevel string) {
	if minLevel == "" {
		minLevel = "WARN"
	}
	c.cfgMu.Lock()
	c.persistEnabled = enabled
	c.persistMinLevel = minLevel
	c.cfgMu.Unlock()
}

func (c *LogCollector) append(entries []LogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf = append(c.buf, entries...)
	if len(c.buf) > logBufferSize {
		c.buf = c.buf[len(c.buf)-logBufferSize:]
	}
}

func (c *LogCollector) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf = c.buf[:0]
}

// Recent returns the last n entries from the in-memory buffer,
// optionally filtered by minimum level and keyword.
func (c *LogCollector) Recent(n int, level, keyword string) []LogEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	minLevel := levelRank(level)
	kw := strings.ToLower(keyword)
	result := make([]LogEntry, 0, n)
	for i := len(c.buf) - 1; i >= 0 && len(result) < n; i-- {
		e := c.buf[i]
		if levelRank(e.Level) > minLevel {
			continue
		}
		if kw != "" && !strings.Contains(strings.ToLower(e.Message), kw) {
			continue
		}
		result = append(result, e)
	}
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func levelRank(level string) int {
	switch level {
	case "PANIC":
		return 0
	case "FATAL":
		return 1
	case "ERROR":
		return 2
	case "WARN":
		return 3
	case "INFO":
		return 4
	case "DEBUG":
		return 5
	case "TRACE":
		return 6
	default:
		return 6
	}
}

func (c *LogCollector) Run(ctx context.Context) {
	for {
		if err := c.connect(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[%s] log collector error: %v, retrying...", c.instance, err)
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

func (c *LogCollector) connect(ctx context.Context) error {
	conn, err := newGRPCConn(c.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := daemon.NewStartedServiceClient(conn)
	authCtx := withAuth(ctx, c.secret)

	stream, err := client.SubscribeLog(authCtx, &emptypb.Empty{})
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
			c.clear()
		}

		entries := make([]LogEntry, 0, len(msg.Messages))
		now := time.Now()
		for _, m := range msg.Messages {
			entries = append(entries, LogEntry{
				Time:    now,
				Level:   m.Level.String(),
				Message: m.Message,
			})
		}
		if len(entries) == 0 {
			continue
		}

		c.append(entries)

		// persist to DB if enabled and level meets threshold
		c.cfgMu.RLock()
		enabled := c.persistEnabled
		minLevel := c.persistMinLevel
		c.cfgMu.RUnlock()

		if enabled && c.db != nil {
			minRank := levelRank(minLevel)
			tsMs := now.UnixMilli()
			for _, e := range entries {
				if levelRank(e.Level) <= minRank {
					if err := store.InsertLog(c.db, c.instance, tsMs, e.Level, e.Message); err != nil {
						log.Printf("[%s] log persist error: %v", c.instance, err)
					}
				}
			}
		}
	}
}
