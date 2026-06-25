import { useMemo } from 'react'
import { Connection } from '@/lib/api'
import { formatBytes, formatTime } from '@/lib/utils'
import { useGeo, geoLabel, GeoInfo } from '@/hooks/useGeo'

type SortField = 'upload' | 'download' | ''
type SortDir = 'asc' | 'desc'

interface Props {
  connections: Connection[]
  total?: number
  page?: number
  limit?: number
  onPageChange?: (page: number) => void
  sortBy?: SortField
  sortDir?: SortDir
  onSortChange?: (field: SortField, dir: SortDir) => void
}

export function ConnectionTable({ connections, total = 0, page = 1, limit = 20, onPageChange, sortBy = '', sortDir = 'desc', onSortChange }: Props) {
  const totalPages = Math.ceil(total / limit)

  const handleSort = (field: SortField) => {
    if (!onSortChange) return
    if (sortBy === field) {
      onSortChange(field, sortDir === 'desc' ? 'asc' : 'desc')
    } else {
      onSortChange(field, 'desc')
    }
  }

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortBy !== field) return <span className="ml-1 opacity-30">⇅</span>
    return <span className="ml-1">{sortDir === 'desc' ? '↓' : '↑'}</span>
  }

  // 收集所有需要查询的 IP
  const ips = useMemo(() =>
    connections.flatMap(c => [c.source_ip, c.dest_ip].filter(Boolean)),
    [connections]
  )
  const geoMap = useGeo(ips)

  return (
    <div>
      <div className="overflow-x-auto rounded-xl border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="text-left px-4 py-3 font-medium text-muted-foreground whitespace-nowrap">时间</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground hidden sm:table-cell whitespace-nowrap">入站</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground whitespace-nowrap">源地址</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground whitespace-nowrap">目标</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground whitespace-nowrap">出站</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground hidden lg:table-cell whitespace-nowrap">规则</th>
              <th className="px-4 py-3 font-medium text-muted-foreground text-right whitespace-nowrap">
                <button
                  onClick={() => handleSort('upload')}
                  className="inline-flex items-center hover:text-foreground transition-colors"
                >
                  上传<SortIcon field="upload" />
                </button>
              </th>
              <th className="px-4 py-3 font-medium text-muted-foreground text-right whitespace-nowrap">
                <button
                  onClick={() => handleSort('download')}
                  className="inline-flex items-center hover:text-foreground transition-colors"
                >
                  下载<SortIcon field="download" />
                </button>
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground hidden xl:table-cell whitespace-nowrap">进程</th>
            </tr>
          </thead>
          <tbody>
            {connections.length === 0 && (
              <tr>
                <td colSpan={9} className="text-center py-12 text-muted-foreground">暂无数据</td>
              </tr>
            )}
            {connections.map(c => (
              <tr key={c.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                <td className="px-4 py-2.5 whitespace-nowrap text-muted-foreground text-xs">
                  {formatTime(c.started_at)}
                </td>
                <td className="px-4 py-2.5 hidden sm:table-cell">
                  <InboundBadge type={c.inbound_type} name={c.inbound} />
                </td>
                <td className="px-4 py-2.5">
                  <AddrCell ip={c.source_ip} port={c.source_port} geo={geoMap[c.source_ip]} />
                </td>
                <td className="px-4 py-2.5 max-w-[180px] sm:max-w-[240px]">
                  <AddrCell host={c.host} ip={c.dest_ip} port={c.dest_port} geo={geoMap[c.dest_ip]} />
                </td>
                <td className="px-4 py-2.5">
                  {c.outbound && (
                    <span className="inline-flex items-center rounded-md bg-accent px-2 py-0.5 text-xs font-medium">
                      {c.outbound}
                    </span>
                  )}
                </td>
                <td className="px-4 py-2.5 max-w-[160px] hidden lg:table-cell">
                  <div className="truncate text-xs text-muted-foreground">{c.rule || '-'}</div>
                </td>
                <td className="px-4 py-2.5 text-right whitespace-nowrap text-xs text-indigo-500">
                  {formatBytes(c.upload)}
                </td>
                <td className="px-4 py-2.5 text-right whitespace-nowrap text-xs text-emerald-500">
                  {formatBytes(c.download)}
                </td>
                <td className="px-4 py-2.5 max-w-[120px] hidden xl:table-cell">
                  <div className="truncate text-xs text-muted-foreground">{c.process_path || '-'}</div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {totalPages > 1 && onPageChange && (
        <div className="flex items-center justify-between mt-4 text-sm">
          <span className="text-muted-foreground">共 {total} 条</span>
          <div className="flex gap-2">
            <button
              disabled={page <= 1}
              onClick={() => onPageChange(page - 1)}
              className="px-3 py-1 rounded-md border hover:bg-accent disabled:opacity-40 disabled:cursor-not-allowed"
            >
              上一页
            </button>
            <span className="px-3 py-1 text-muted-foreground">{page} / {totalPages}</span>
            <button
              disabled={page >= totalPages}
              onClick={() => onPageChange(page + 1)}
              className="px-3 py-1 rounded-md border hover:bg-accent disabled:opacity-40 disabled:cursor-not-allowed"
            >
              下一页
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

function AddrCell({ host, ip, port, geo }: { host?: string; ip?: string; port?: number; geo?: GeoInfo }) {
  const hasHost = host && host !== ip
  const ipPort = [ip, port ? String(port) : ''].filter(Boolean).join(':')
  const label = geoLabel(geo)

  return (
    <div className="font-mono text-xs leading-tight">
      {hasHost ? (
        <>
          <div className="truncate max-w-[220px]">{host}</div>
          <div className="text-muted-foreground truncate max-w-[260px]">
            {ipPort && `(${ipPort}) `}
            {label && <span className="not-italic">{label}</span>}
          </div>
        </>
      ) : (
        <>
          <div className="truncate max-w-[220px]">{ipPort || '-'}</div>
          {label && <div className="text-muted-foreground not-italic">{label}</div>}
        </>
      )}
    </div>
  )
}

function InboundBadge({ type, name }: { type: string; name: string }) {
  const colors: Record<string, string> = {
    tun: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    mixed: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
    vmess: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    vless: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    trojan: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    socks: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400',
    naive: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400',
  }
  const color = colors[type] || 'bg-muted text-muted-foreground'
  const label = name ? `${type}/${name}` : type
  return (
    <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ${color}`}>
      {label || '-'}
    </span>
  )
}
