package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type Connection struct {
	ID          string  `json:"id"`
	Instance    string  `json:"instance"`
	Network     string  `json:"network"`
	Inbound     string  `json:"inbound"`
	InboundType string  `json:"inbound_type"`
	Outbound    string  `json:"outbound"`
	OutboundType string `json:"outbound_type"`
	SourceIP    string  `json:"source_ip"`
	SourcePort  int     `json:"source_port"`
	DestIP      string  `json:"dest_ip"`
	DestPort    int     `json:"dest_port"`
	Host        string  `json:"host"`
	ProcessPath string  `json:"process_path"`
	Rule        string  `json:"rule"`
	Chains      string  `json:"chains"`
	Upload      int64   `json:"upload"`
	Download    int64   `json:"download"`
	StartedAt   int64   `json:"started_at"`
	ClosedAt    *int64  `json:"closed_at"`
}

func UpsertConnection(db *sql.DB, c *Connection) error {
	_, err := db.Exec(`
INSERT INTO connections(id, instance, network, inbound, inbound_type, outbound, outbound_type,
    source_ip, source_port, dest_ip, dest_port, host, process_path, rule, chains,
    upload, download, started_at, closed_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    upload = excluded.upload,
    download = excluded.download,
    closed_at = excluded.closed_at`,
		c.ID, c.Instance, c.Network, c.Inbound, c.InboundType, c.Outbound, c.OutboundType,
		c.SourceIP, c.SourcePort, c.DestIP, c.DestPort, c.Host, c.ProcessPath, c.Rule, c.Chains,
		c.Upload, c.Download, c.StartedAt, c.ClosedAt,
	)
	return err
}

func CloseConnection(db *sql.DB, id string, closedAt int64, upload, download int64) error {
	_, err := db.Exec(
		`UPDATE connections SET closed_at = ?, upload = ?, download = ? WHERE id = ?`,
		closedAt, upload, download, id,
	)
	return err
}

type ConnectionFilter struct {
	Instance     string
	Inbound      string
	InboundType  string
	Outbound     string
	Search       string // 匹配 host 或 dest_ip
	SourceSearch string // 匹配 source_ip
	Rule         string
	From         int64
	To           int64
	ActiveOnly   bool
	Page         int
	Limit        int
}

func QueryConnections(db *sql.DB, f ConnectionFilter) ([]Connection, int, error) {
	where, args := buildConnectionWhere(f)

	var total int
	countSQL := `SELECT COUNT(*) FROM connections` + where
	if err := db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	querySQL := `SELECT id, instance, network, inbound, inbound_type, outbound, outbound_type,
		source_ip, source_port, dest_ip, dest_port, host, process_path, rule, chains,
		upload, download, started_at, closed_at
		FROM connections` + where + ` ORDER BY started_at DESC LIMIT ? OFFSET ?`
	args = append(args, f.Limit, offset)

	rows, err := db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var conns []Connection
	for rows.Next() {
		var c Connection
		if err := rows.Scan(
			&c.ID, &c.Instance, &c.Network, &c.Inbound, &c.InboundType, &c.Outbound, &c.OutboundType,
			&c.SourceIP, &c.SourcePort, &c.DestIP, &c.DestPort, &c.Host, &c.ProcessPath, &c.Rule, &c.Chains,
			&c.Upload, &c.Download, &c.StartedAt, &c.ClosedAt,
		); err != nil {
			return nil, 0, err
		}
		conns = append(conns, c)
	}
	return conns, total, rows.Err()
}

func buildConnectionWhere(f ConnectionFilter) (string, []any) {
	var conds []string
	var args []any

	if f.Instance != "" {
		conds = append(conds, "instance = ?")
		args = append(args, f.Instance)
	}
	if f.Inbound != "" {
		conds = append(conds, "inbound = ?")
		args = append(args, f.Inbound)
	}
	if f.InboundType != "" {
		conds = append(conds, "inbound_type = ?")
		args = append(args, f.InboundType)
	}
	if f.Outbound != "" {
		conds = append(conds, "outbound = ?")
		args = append(args, f.Outbound)
	}
	if f.Search != "" {
		conds = append(conds, "(host LIKE ? OR dest_ip LIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like)
	}
	if f.SourceSearch != "" {
		conds = append(conds, "source_ip LIKE ?")
		args = append(args, "%"+f.SourceSearch+"%")
	}
	if f.Rule != "" {
		conds = append(conds, "rule LIKE ?")
		args = append(args, "%"+f.Rule+"%")
	}
	if f.From > 0 {
		conds = append(conds, "started_at >= ?")
		args = append(args, f.From)
	}
	if f.To > 0 {
		conds = append(conds, "started_at <= ?")
		args = append(args, f.To)
	}
	if f.ActiveOnly {
		conds = append(conds, "closed_at IS NULL")
	}

	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

type TopDomain struct {
	Host     string `json:"host"`
	Count    int    `json:"count"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

func QueryTopDomains(db *sql.DB, instance string, from int64, limit int) ([]TopDomain, error) {
	rows, err := db.Query(`
SELECT host, COUNT(*) as cnt, SUM(upload), SUM(download)
FROM connections
WHERE instance = ? AND started_at >= ? AND host != ''
GROUP BY host
ORDER BY cnt DESC
LIMIT ?`, instance, from, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopDomain
	for rows.Next() {
		var d TopDomain
		if err := rows.Scan(&d.Host, &d.Count, &d.Upload, &d.Download); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

type TopOutbound struct {
	Outbound string `json:"outbound"`
	Count    int    `json:"count"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

func QueryTopOutbounds(db *sql.DB, instance string, from int64, limit int) ([]TopOutbound, error) {
	rows, err := db.Query(`
SELECT outbound, COUNT(*) as cnt, SUM(upload), SUM(download)
FROM connections
WHERE instance = ? AND started_at >= ? AND outbound != ''
GROUP BY outbound
ORDER BY cnt DESC
LIMIT ?`, instance, from, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopOutbound
	for rows.Next() {
		var o TopOutbound
		if err := rows.Scan(&o.Outbound, &o.Count, &o.Upload, &o.Download); err != nil {
			return nil, err
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func QueryInbounds(db *sql.DB, instance string) ([]string, error) {
	return queryDistinct(db, "inbound", instance)
}

func QueryOutbounds(db *sql.DB, instance string) ([]string, error) {
	return queryDistinct(db, "outbound", instance)
}

func queryDistinct(db *sql.DB, col, instance string) ([]string, error) {
	q := fmt.Sprintf(`SELECT DISTINCT %s FROM connections WHERE instance = ? AND %s != '' ORDER BY %s`, col, col, col)
	rows, err := db.Query(q, instance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func DeleteOldConnections(db *sql.DB, before int64) error {
	_, err := db.Exec(`DELETE FROM connections WHERE started_at < ? AND closed_at IS NOT NULL`, before)
	return err
}
