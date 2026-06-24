package collector

import (
	"context"
	"database/sql"
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/daemon"

	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type TrafficCollector struct {
	instance string
	apiURL   string
	secret   string
	db       *sql.DB

	mu          sync.RWMutex
	currentUp   int64
	currentDown int64
}

func NewTrafficCollector(instance, apiURL, secret string, db *sql.DB) *TrafficCollector {
	return &TrafficCollector{instance: instance, apiURL: apiURL, secret: secret, db: db}
}

func (c *TrafficCollector) Current() (up, down int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentUp, c.currentDown
}

func (c *TrafficCollector) Run(ctx context.Context) {
	for {
		if err := c.connect(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[%s] traffic collector error: %v, retrying...", c.instance, err)
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

func (c *TrafficCollector) connect(ctx context.Context) error {
	conn, err := newGRPCConn(c.apiURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := daemon.NewStartedServiceClient(conn)
	authCtx := withAuth(ctx, c.secret)

	stream, err := client.SubscribeStatus(authCtx, &daemon.SubscribeStatusRequest{Interval: int64(time.Second)})
	if err != nil {
		return err
	}

	log.Printf("[%s] traffic collector connected", c.instance)

	var (
		minBucket int64
		minUpSum  int64
		minDnSum  int64
		minCount  int
	)
	flush := func(bucket int64) {
		if minCount == 0 {
			return
		}
		store.InsertTrafficStat(c.db, c.instance, bucket, minUpSum/int64(minCount), minDnSum/int64(minCount))
		minUpSum, minDnSum, minCount = 0, 0, 0
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
		c.currentUp = msg.Uplink
		c.currentDown = msg.Downlink
		c.mu.Unlock()

		now := time.Now().UnixMilli()
		bucket := now - (now % 60000) // 按分钟聚合，单位毫秒
		if minBucket == 0 {
			minBucket = bucket
		}
		if bucket != minBucket {
			flush(minBucket)
			minBucket = bucket
		}
		minUpSum += msg.Uplink
		minDnSum += msg.Downlink
		minCount++
	}
}

var retryCount int
var retryMu sync.Mutex

func reconnectDelay() time.Duration {
	retryMu.Lock()
	retryCount++
	n := retryCount
	retryMu.Unlock()
	d := time.Duration(math.Min(float64(n*n), 30)) * time.Second
	if d < time.Second {
		d = time.Second
	}
	return d
}
