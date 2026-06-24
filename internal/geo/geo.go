package geo

import (
	"net"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

type Info struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Region      string `json:"region,omitempty"` // 省/州
	City        string `json:"city,omitempty"`
}

// Label 返回适合展示的完整地理位置字符串
func (i Info) Label() string {
	parts := []string{}
	if i.CountryName != "" {
		parts = append(parts, i.CountryName)
	}
	if i.Region != "" && i.Region != i.CountryName {
		parts = append(parts, i.Region)
	}
	if i.City != "" && i.City != i.Region {
		parts = append(parts, i.City)
	}
	return strings.Join(parts, " · ")
}

var (
	db  *geoip2.Reader
	mu  sync.RWMutex
	once sync.Once
)

func Init(dbPath string) error {
	var err error
	once.Do(func() {
		db, err = geoip2.Open(dbPath)
	})
	return err
}

// Reinit 允许在运行时更换 GeoIP 数据库路径（设置页修改路径时调用）。
func Reinit(dbPath string) error {
	r, err := geoip2.Open(dbPath)
	if err != nil {
		return err
	}
	mu.Lock()
	if db != nil {
		db.Close()
	}
	db = r
	mu.Unlock()
	return nil
}

func Lookup(ipStr string) Info {
	mu.RLock()
	r := db
	mu.RUnlock()
	if r == nil || ipStr == "" {
		return Info{}
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Info{}
	}
	record, err := r.City(ip)
	if err != nil {
		return Info{}
	}

	name := func(m map[string]string) string {
		if v := m["zh-CN"]; v != "" {
			return v
		}
		return m["en"]
	}

	info := Info{
		CountryCode: record.Country.IsoCode,
		CountryName: name(record.Country.Names),
		City:        name(record.City.Names),
	}
	if len(record.Subdivisions) > 0 {
		info.Region = name(record.Subdivisions[0].Names)
	}
	return info
}
