package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/laurdawn/sing-box-watcher/internal/geo"
	"github.com/laurdawn/sing-box-watcher/internal/store"
)

type RegionStat struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Region      string `json:"region,omitempty"`
	Flag        string `json:"flag"`
	Count       int    `json:"count"`
	Upload      int64  `json:"upload"`
	Download    int64  `json:"download"`
	IPs         int    `json:"ips"` // 独立 IP 数
}

func (s *Server) handleTopSourceIPs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	instance := q.Get("instance")
	hours, _ := strconv.Atoi(q.Get("hours"))
	if hours <= 0 {
		hours = 24
	}
	from := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()

	result, err := store.QuerySourceIPs(s.db, instance, from, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []store.SourceIPStat{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleSourceRegions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	instance := q.Get("instance")
	hours, _ := strconv.Atoi(q.Get("hours"))
	if hours <= 0 {
		hours = 24
	}
	from := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()

	ipStats, err := store.QuerySourceIPs(s.db, instance, from, 2000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 按国家聚合
	type key struct{ code, name, region string }
	grouped := make(map[key]*RegionStat)

	for _, s := range ipStats {
		info := geo.Lookup(s.SourceIP)
		if info.CountryCode == "" {
			info.CountryCode = "XX"
			info.CountryName = "未知"
		}
		k := key{info.CountryCode, info.CountryName, ""}
		if _, ok := grouped[k]; !ok {
			grouped[k] = &RegionStat{
				CountryCode: info.CountryCode,
				CountryName: info.CountryName,
				Flag:        countryFlagEmoji(info.CountryCode),
			}
		}
		grouped[k].Count += s.Count
		grouped[k].Upload += s.Upload
		grouped[k].Download += s.Download
		grouped[k].IPs++
	}

	result := make([]RegionStat, 0, len(grouped))
	for _, v := range grouped {
		result = append(result, *v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	writeJSON(w, http.StatusOK, result)
}

func countryFlagEmoji(code string) string {
	if len(code) != 2 {
		return ""
	}
	r := []rune(code)
	return string([]rune{
		rune(0x1F1E6 + r[0] - 'A'),
		rune(0x1F1E6 + r[1] - 'A'),
	})
}
