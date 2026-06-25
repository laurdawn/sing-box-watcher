import { Activity, ArrowDown, ArrowUp, Wifi, Search, Clock, AlertTriangle } from 'lucide-react'
import { StatCard } from '@/components/StatCard'
import { TrafficChart } from '@/components/TrafficChart'
import { ConnectionTable } from '@/components/ConnectionTable'
import { useWsTraffic } from '@/hooks/useWsTraffic'
import { InstanceStats, ServiceInfo, api, Connection } from '@/lib/api'
import { formatSpeed } from '@/lib/utils'
import { useEffect, useState, useMemo } from 'react'

interface Props {
  instance: string
  instances: InstanceStats[]
}

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
}

export function Dashboard({ instance, instances }: Props) {
  const liveTraffic = useWsTraffic()
  const [activeConns, setActiveConns] = useState<Connection[]>([])
  const [sourceFilter, setSourceFilter] = useState('')
  const [destFilter, setDestFilter] = useState('')
  const [serviceInfo, setServiceInfo] = useState<ServiceInfo | null>(null)

  const live = liveTraffic.find(t => t.instance === instance)
  const stats = instances.find(i => i.name === instance)

  useEffect(() => {
    if (!instance) return
    const load = () => {
      api.activeConnections(instance)
        .then(r => setActiveConns(r.connections || []))
        .catch(() => {})
    }
    load()
    const t = setInterval(load, 2000)
    return () => clearInterval(t)
  }, [instance])

  useEffect(() => {
    if (!instance) return
    const load = () => {
      api.serviceInfo(instance).then(setServiceInfo).catch(() => {})
    }
    load()
    const t = setInterval(load, 10000)
    return () => clearInterval(t)
  }, [instance])

  // 客户端过滤（数据已在内存，不走 API）
  const filtered = useMemo(() => {
    let conns = activeConns
    if (sourceFilter) {
      const kw = sourceFilter.toLowerCase()
      conns = conns.filter(c => c.source_ip?.toLowerCase().includes(kw))
    }
    if (destFilter) {
      const kw = destFilter.toLowerCase()
      conns = conns.filter(c =>
        c.host?.toLowerCase().includes(kw) ||
        c.dest_ip?.toLowerCase().includes(kw)
      )
    }
    return conns
  }, [activeConns, sourceFilter, destFilter])

  return (
    <div className="space-y-6">
      {/* 废弃配置警告 */}
      {serviceInfo && serviceInfo.deprecated_warnings.length > 0 && (
        <div className="flex items-start gap-2 px-4 py-3 rounded-lg border border-amber-200 bg-amber-50 dark:bg-amber-900/20 dark:border-amber-800 text-sm text-amber-800 dark:text-amber-300">
          <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
          <div>
            <div className="font-medium mb-1">配置项已废弃</div>
            {serviceInfo.deprecated_warnings.map((w, i) => <div key={i} className="text-xs">{w}</div>)}
          </div>
        </div>
      )}

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <StatCard
          label="上传速率"
          value={formatSpeed(live?.up ?? 0)}
          icon={<ArrowUp className="w-4 h-4 text-indigo-500" />}
        />
        <StatCard
          label="下载速率"
          value={formatSpeed(live?.down ?? 0)}
          icon={<ArrowDown className="w-4 h-4 text-emerald-500" />}
        />
        <StatCard
          label="活跃连接"
          value={String(stats?.active_connections ?? activeConns.length)}
          icon={<Wifi className="w-4 h-4" />}
        />
        <StatCard
          label="服务状态"
          value={stats?.status || (stats?.online !== false ? '在线' : '离线')}
          icon={<Activity className={`w-4 h-4 ${stats?.online ? 'text-emerald-500' : 'text-red-500'}`} />}
        />
        <StatCard
          label="版本"
          value={serviceInfo?.version || '-'}
          icon={<Activity className="w-4 h-4 text-muted-foreground" />}
        />
        <StatCard
          label="运行时长"
          value={serviceInfo ? formatUptime(serviceInfo.uptime_seconds) : '-'}
          icon={<Clock className="w-4 h-4 text-muted-foreground" />}
        />
      </div>

      <TrafficChart instance={instance} />

      <div>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-3">
          <h3 className="font-medium text-sm text-muted-foreground">
            活跃连接（实时）
            {(sourceFilter || destFilter) && (
              <span className="ml-2 text-xs text-foreground">
                {filtered.length} / {activeConns.length}
              </span>
            )}
          </h3>
          <div className="flex items-center gap-2 flex-wrap">
            <div className="relative">
              <Search className="absolute left-2.5 top-2 w-3.5 h-3.5 text-muted-foreground" />
              <input
                value={sourceFilter}
                onChange={e => setSourceFilter(e.target.value)}
                placeholder="过滤源 IP"
                className="h-8 pl-7 pr-3 rounded-md border bg-background text-xs focus:outline-none focus:ring-1 focus:ring-primary w-28 sm:w-36"
              />
            </div>
            <div className="relative">
              <Search className="absolute left-2.5 top-2 w-3.5 h-3.5 text-muted-foreground" />
              <input
                value={destFilter}
                onChange={e => setDestFilter(e.target.value)}
                placeholder="过滤目标域名/IP"
                className="h-8 pl-7 pr-3 rounded-md border bg-background text-xs focus:outline-none focus:ring-1 focus:ring-primary w-36 sm:w-44"
              />
            </div>
          </div>
        </div>
        <ConnectionTable
          connections={filtered.slice(0, 50)}
          total={filtered.length}
        />
      </div>
    </div>
  )
}
