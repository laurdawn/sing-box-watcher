package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type apiClient struct {
	baseURL string
	http    http.Client
}

func (c *apiClient) get(path string) (string, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

func (c *apiClient) post(path string, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}
	resp, err := c.http.Post(c.baseURL+path, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return string(body), nil
}

func ok(text string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(text), nil
}

func fail(err error) (*mcp.CallToolResult, error) {
	return nil, err
}

// --- tool handlers ---

func (c *apiClient) listInstances(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body, err := c.get("/api/instances")
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) getServiceInfo(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	body, err := c.get("/api/service/info?instance=" + instance)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) queryTraffic(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	from := req.GetInt("from", 0)
	to := req.GetInt("to", 0)
	path := "/api/traffic?instance=" + instance
	if from > 0 {
		path += fmt.Sprintf("&from=%d", from)
	}
	if to > 0 {
		path += fmt.Sprintf("&to=%d", to)
	}
	body, err := c.get(path)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) queryConnections(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	path := "/api/connections?instance=" + instance
	if v := req.GetString("search", ""); v != "" {
		path += "&search=" + v
	}
	if v := req.GetString("inbound", ""); v != "" {
		path += "&inbound=" + v
	}
	if v := req.GetString("outbound", ""); v != "" {
		path += "&outbound=" + v
	}
	if v := req.GetString("rule", ""); v != "" {
		path += "&rule=" + v
	}
	if v := req.GetString("sort_by", ""); v != "" {
		path += "&sort_by=" + v
	}
	if v := req.GetString("sort_dir", ""); v != "" {
		path += "&sort_dir=" + v
	}
	page := req.GetInt("page", 1)
	limit := req.GetInt("limit", 20)
	path += fmt.Sprintf("&page=%d&limit=%d", page, limit)
	body, err := c.get(path)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) getActiveConnections(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	body, err := c.get("/api/connections/active?instance=" + instance)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) getTopDomains(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	hours := req.GetInt("hours", 24)
	body, err := c.get(fmt.Sprintf("/api/stats/top-domains?instance=%s&hours=%d", instance, hours))
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) getTopOutbounds(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	hours := req.GetInt("hours", 24)
	body, err := c.get(fmt.Sprintf("/api/stats/top-outbounds?instance=%s&hours=%d", instance, hours))
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) getSourceRegions(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	hours := req.GetInt("hours", 24)
	body, err := c.get(fmt.Sprintf("/api/stats/source-regions?instance=%s&hours=%d", instance, hours))
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) getTopSourceIPs(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	hours := req.GetInt("hours", 24)
	body, err := c.get(fmt.Sprintf("/api/stats/top-source-ips?instance=%s&hours=%d", instance, hours))
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) listProxyGroups(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	instance := req.GetString("instance", "")
	body, err := c.get("/api/groups?instance=" + instance)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) selectOutbound(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]string{
		"instance":     req.GetString("instance", ""),
		"group_tag":    req.GetString("group_tag", ""),
		"outbound_tag": req.GetString("outbound_tag", ""),
	}
	body, err := c.post("/api/groups/select", payload)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

func (c *apiClient) lookupGeo(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload := map[string]string{"ip": req.GetString("ip", "")}
	body, err := c.post("/api/geo/lookup", payload)
	if err != nil {
		return fail(err)
	}
	return ok(body)
}

