package store

import (
	"database/sql"
)

type TrafficPoint struct {
	TS       int64  `json:"ts"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

func InsertTrafficStat(db *sql.DB, instance string, ts, upload, download int64) error {
	_, err := db.Exec(
		`INSERT INTO traffic_stats(instance, ts, upload, download) VALUES(?, ?, ?, ?)`,
		instance, ts, upload, download,
	)
	return err
}

func QueryTraffic(db *sql.DB, instance string, from, to int64) ([]TrafficPoint, error) {
	rows, err := db.Query(
		`SELECT ts, upload, download FROM traffic_stats
		 WHERE instance = ? AND ts >= ? AND ts <= ?
		 ORDER BY ts ASC`,
		instance, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TrafficPoint
	for rows.Next() {
		var p TrafficPoint
		if err := rows.Scan(&p.TS, &p.Upload, &p.Download); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func DeleteOldTraffic(db *sql.DB, before int64) error {
	_, err := db.Exec(`DELETE FROM traffic_stats WHERE ts < ?`, before)
	return err
}
