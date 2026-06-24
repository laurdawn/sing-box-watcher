package store

import "database/sql"

type SourceIPStat struct {
	SourceIP string `json:"source_ip"`
	Count    int    `json:"count"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

// QuerySourceIPs 返回指定时间范围内所有唯一源 IP 及其连接数和流量。
func QuerySourceIPs(db *sql.DB, instance string, from int64, limit int) ([]SourceIPStat, error) {
	rows, err := db.Query(`
SELECT source_ip, COUNT(*) as cnt, SUM(upload), SUM(download)
FROM connections
WHERE instance = ? AND started_at >= ? AND source_ip != ''
GROUP BY source_ip
ORDER BY cnt DESC
LIMIT ?`, instance, from, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SourceIPStat
	for rows.Next() {
		var s SourceIPStat
		if err := rows.Scan(&s.SourceIP, &s.Count, &s.Upload, &s.Download); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
