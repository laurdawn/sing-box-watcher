package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type LogFilter struct {
	Instance string
	Level    string // minimum level filter
	Keyword  string
	From     int64 // Unix seconds
	To       int64 // Unix seconds
	Limit    int
}

func InsertLog(db *sql.DB, instance string, ts int64, level, message string) error {
	_, err := db.Exec(
		`INSERT INTO logs(instance, ts, level, message) VALUES(?, ?, ?, ?)`,
		instance, ts, level, message,
	)
	return err
}

func QueryLogs(db *sql.DB, f LogFilter) ([]LogEntry, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	where := []string{"instance = ?"}
	args := []any{f.Instance}

	if f.Level != "" {
		// collect all levels with rank <= requested level
		rank := logLevelRank(f.Level)
		levels := []string{}
		for _, l := range []string{"PANIC", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"} {
			if logLevelRank(l) <= rank {
				levels = append(levels, "'"+l+"'")
			}
		}
		if len(levels) > 0 {
			where = append(where, "level IN ("+strings.Join(levels, ",")+")")
		}
	}
	if f.From > 0 {
		where = append(where, "ts >= ?")
		args = append(args, f.From*1000)
	}
	if f.To > 0 {
		where = append(where, "ts <= ?")
		args = append(args, f.To*1000)
	}
	if f.Keyword != "" {
		where = append(where, "message LIKE ?")
		args = append(args, "%"+f.Keyword+"%")
	}

	q := fmt.Sprintf(
		`SELECT ts, level, message FROM logs WHERE %s ORDER BY ts DESC LIMIT %d`,
		strings.Join(where, " AND "), limit,
	)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]LogEntry, 0, limit)
	for rows.Next() {
		var ts int64
		var level, message string
		if err := rows.Scan(&ts, &level, &message); err != nil {
			return nil, err
		}
		result = append(result, LogEntry{
			TSMillis: ts,
			Level:    level,
			Message:  message,
		})
	}
	// reverse to chronological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, rows.Err()
}

func DeleteOldLogs(db *sql.DB, before int64) error {
	_, err := db.Exec(`DELETE FROM logs WHERE ts < ?`, before)
	return err
}

func logLevelRank(level string) int {
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

// LogEntry is the store-layer representation (TSMillis instead of time.Time).
type LogEntry struct {
	TSMillis int64  `json:"ts"`
	Level    string `json:"level"`
	Message  string `json:"message"`
}
