export interface InstanceStats {
  name: string
  current_up: number
  current_down: number
  active_connections: number
  online: boolean
  status: string
}

export interface TrafficPoint {
  ts: number
  upload: number
  download: number
}

export interface Connection {
  id: string
  instance: string
  network: string
  inbound: string
  inbound_type: string
  outbound: string
  outbound_type: string
  source_ip: string
  source_port: number
  dest_ip: string
  dest_port: number
  host: string
  process_path: string
  rule: string
  chains: string
  upload: number
  download: number
  started_at: number
  closed_at: number | null
}

export interface ConnectionsResponse {
  total: number
  page: number
  limit: number
  connections: Connection[]
}

export interface TopDomain {
  host: string
  count: number
  upload: number
  download: number
}

export interface TopOutbound {
  outbound: string
  count: number
  upload: number
  download: number
}

const base = ''

async function get<T>(path: string, params?: Record<string, string | number>): Promise<T> {
  const url = new URL(base + path, window.location.origin)
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') url.searchParams.set(k, String(v))
    })
  }
  const res = await fetch(url.toString())
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export interface ServiceInfo {
  version: string
  api_version: number
  started_at_ms: number
  uptime_seconds: number
  status: string
  online: boolean
  deprecated_warnings: string[]
}

export interface GroupItem {
  tag: string
  type: string
  urlTestTime: number
  urlTestDelay: number
}

export interface Group {
  tag: string
  type: string
  selectable: boolean
  selected: string
  isExpand: boolean
  items: GroupItem[]
}

export interface GroupsSnapshot {
  groups: Group[]
  updatedAt: number
}

export interface LogEntry {
  level: string
  message: string
}

export interface RegionStat {
  country_code: string
  country_name: string
  flag: string
  count: number
  upload: number
  download: number
  ips: number
}

export interface SourceIPStat {
  source_ip: string
  count: number
  upload: number
  download: number
}

export interface GeoInfo {
  country_code: string
  country_name: string
  region?: string
  city?: string
}

export const api = {
  instances: () => get<InstanceStats[]>('/api/instances'),

  traffic: (instance: string, from: number, to: number) =>
    get<{ instance: string; points: TrafficPoint[] }>('/api/traffic', { instance, from, to }),

  connections: (params: {
    instance?: string
    inbound?: string
    inbound_type?: string
    outbound?: string
    source?: string
    search?: string
    rule?: string
    from?: number
    to?: number
    page?: number
    limit?: number
    sort_by?: string
    sort_dir?: string
  }) => get<ConnectionsResponse>('/api/connections', params as Record<string, string | number>),

  activeConnections: (instance: string) =>
    get<{ total: number; connections: Connection[] }>('/api/connections/active', { instance }),

  inbounds: (instance: string) => get<string[]>('/api/connections/inbounds', { instance }),
  outbounds: (instance: string) => get<string[]>('/api/connections/outbounds', { instance }),

  topDomains: (instance: string, hours = 24) =>
    get<TopDomain[]>('/api/stats/top-domains', { instance, hours }),

  topOutbounds: (instance: string, hours = 24) =>
    get<TopOutbound[]>('/api/stats/top-outbounds', { instance, hours }),

  sourceRegions: (instance: string, hours = 24) =>
    get<RegionStat[]>('/api/stats/source-regions', { instance, hours }),

  topSourceIPs: (instance: string, hours = 24) =>
    get<SourceIPStat[]>('/api/stats/top-source-ips', { instance, hours }),

  serviceInfo: (instance: string) => get<ServiceInfo>('/api/service/info', { instance }),

  groups: (instance: string) => get<GroupsSnapshot>('/api/groups', { instance }),

  selectOutbound: (instance: string, groupTag: string, outboundTag: string) =>
    fetch('/api/groups/select', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ instance, group_tag: groupTag, outbound_tag: outboundTag }),
    }).then(r => r.json()),

  urlTest: (instance: string, outboundTag: string) =>
    fetch('/api/groups/urltest', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ instance, outbound_tag: outboundTag }),
    }).then(r => r.json()),

  geoLookup: (ips: string[]): Promise<Record<string, GeoInfo>> =>
    fetch('/api/geo/lookup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(ips),
    }).then(r => r.json()),
}
