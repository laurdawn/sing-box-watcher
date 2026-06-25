package mcp

import (
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewHandler creates an MCP Streamable HTTP handler that proxies to the local API.
// baseURL is the internal address of the watcher HTTP server (e.g. "http://localhost:8080").
// internalToken is used to authenticate internal API calls, bypassing cookie-based auth.
func NewHandler(baseURL, internalToken string) http.Handler {
	s := server.NewMCPServer(
		"sing-box-watcher",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithInstructions(
			"Analyze sing-box proxy traffic, connections, and proxy groups. "+
				"Use list_instances first to discover available instances, then query data with other tools.",
		),
	)
	registerTools(s, baseURL, internalToken)
	return server.NewStreamableHTTPServer(s)
}

func registerTools(s *server.MCPServer, baseURL, internalToken string) {
	c := &apiClient{baseURL: baseURL, internalToken: internalToken}

	s.AddTool(mcp.NewTool("list_instances",
		mcp.WithDescription("List all monitored sing-box instances with their online status and current traffic."),
	), c.listInstances)

	s.AddTool(mcp.NewTool("get_service_info",
		mcp.WithDescription("Get sing-box version, uptime, and service status for an instance."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
	), c.getServiceInfo)

	s.AddTool(mcp.NewTool("query_traffic",
		mcp.WithDescription("Query historical traffic data points for an instance within a time range."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithNumber("from", mcp.Description("Start time as Unix timestamp (milliseconds)")),
		mcp.WithNumber("to", mcp.Description("End time as Unix timestamp (milliseconds)")),
	), c.queryTraffic)

	s.AddTool(mcp.NewTool("query_connections",
		mcp.WithDescription("Query connection records with optional filtering. Returns paginated results."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithString("search", mcp.Description("Filter by host or destination IP")),
		mcp.WithString("inbound", mcp.Description("Filter by inbound name")),
		mcp.WithString("outbound", mcp.Description("Filter by outbound name")),
		mcp.WithString("rule", mcp.Description("Filter by rule name")),
		mcp.WithNumber("page", mcp.Description("Page number (default 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page (default 20, max 100)")),
		mcp.WithString("sort_by", mcp.Description("Sort field: started_at, upload, download, host")),
		mcp.WithString("sort_dir", mcp.Description("Sort direction: asc, desc")),
	), c.queryConnections)

	s.AddTool(mcp.NewTool("get_active_connections",
		mcp.WithDescription("Get currently active (open) connections for an instance."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
	), c.getActiveConnections)

	s.AddTool(mcp.NewTool("get_top_domains",
		mcp.WithDescription("Get the most frequently accessed domains in the past N hours."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithNumber("hours", mcp.Description("Look-back window in hours (default 24)")),
	), c.getTopDomains)

	s.AddTool(mcp.NewTool("get_top_outbounds",
		mcp.WithDescription("Get outbound proxies ranked by connection count and traffic in the past N hours."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithNumber("hours", mcp.Description("Look-back window in hours (default 24)")),
	), c.getTopOutbounds)

	s.AddTool(mcp.NewTool("get_source_regions",
		mcp.WithDescription("Get source IP geographic distribution (country/city breakdown) in the past N hours."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithNumber("hours", mcp.Description("Look-back window in hours (default 24)")),
	), c.getSourceRegions)

	s.AddTool(mcp.NewTool("get_top_source_ips",
		mcp.WithDescription("Get source IPs ranked by connection count in the past N hours."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithNumber("hours", mcp.Description("Look-back window in hours (default 24)")),
	), c.getTopSourceIPs)

	s.AddTool(mcp.NewTool("list_proxy_groups",
		mcp.WithDescription("List all proxy groups with their outbounds and current latency."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
	), c.listProxyGroups)

	s.AddTool(mcp.NewTool("select_outbound",
		mcp.WithDescription("Switch the active outbound for a proxy group."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithString("group_tag", mcp.Description("Proxy group tag"), mcp.Required()),
		mcp.WithString("outbound_tag", mcp.Description("Target outbound tag to select"), mcp.Required()),
	), c.selectOutbound)

	s.AddTool(mcp.NewTool("get_recent_logs",
		mcp.WithDescription("Get log entries from sing-box. Without from/to returns recent in-memory entries; with from/to queries persisted history (requires log persistence enabled in settings)."),
		mcp.WithString("instance", mcp.Description("Instance name"), mcp.Required()),
		mcp.WithNumber("n", mcp.Description("Number of recent entries to return (default 100, max 500, used without from/to)")),
		mcp.WithNumber("limit", mcp.Description("Max entries for historical query (default 200, max 1000, used with from/to)")),
		mcp.WithNumber("from", mcp.Description("Start time as Unix timestamp (seconds), enables historical DB query")),
		mcp.WithNumber("to", mcp.Description("End time as Unix timestamp (seconds)")),
		mcp.WithString("level", mcp.Description("Minimum log level filter: ERROR, WARN, INFO, DEBUG, TRACE (default: all)")),
		mcp.WithString("q", mcp.Description("Keyword filter, case-insensitive substring match on message")),
	), c.getRecentLogs)

	s.AddTool(mcp.NewTool("lookup_geo",
		mcp.WithDescription("Look up geographic location (country, city, coordinates) for an IP address."),
		mcp.WithString("ip", mcp.Description("IP address to look up"), mcp.Required()),
	), c.lookupGeo)
}
