import { useEffect, useState } from 'react'
import { api, RegionStat, SourceIPStat } from '@/lib/api'
import { SourceRegionsChart } from '@/components/SourceRegionsChart'
import { formatBytes } from '@/lib/utils'
import { cn } from '@/lib/utils'

interface Props {
  instance: string
}

type Hours = 1 | 6 | 24 | 168

const HOUR_OPTIONS: { label: string; value: Hours }[] = [
  { label: '1小时', value: 1 },
  { label: '6小时', value: 6 },
  { label: '24小时', value: 24 },
  { label: '7天', value: 168 },
]

export function Analysis({ instance }: Props) {
  const [hours, setHours] = useState<Hours>(24)
  const [regions, setRegions] = useState<RegionStat[]>([])
  const [sourceIPs, setSourceIPs] = useState<SourceIPStat[]>([])

  useEffect(() => {
    if (!instance) return
    api.sourceRegions(instance, hours).then(setRegions).catch(() => setRegions([]))
    api.topSourceIPs(instance, hours).then(setSourceIPs).catch(() => setSourceIPs([]))
  }, [instance, hours])

  return (
    <div className="space-y-5">
      {/* 时间范围 */}
      <div className="flex items-center gap-1">
        {HOUR_OPTIONS.map(o => (
          <button
            key={o.value}
            onClick={() => setHours(o.value)}
            className={cn(
              'px-3 py-1.5 text-xs rounded-lg font-medium transition-colors',
              hours === o.value
                ? 'bg-blue-600 text-white'
                : 'text-muted-foreground hover:bg-accent border'
            )}
          >
            {o.label}
          </button>
        ))}
      </div>

      {/* 来源地区 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <SourceRegionsChart data={regions} />
        <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
          <div className="px-5 py-3.5 border-b">
            <h3 className="font-semibold text-sm">来源地区详情</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/40">
                  <th className="text-left px-4 py-2.5 font-medium text-xs text-muted-foreground">#</th>
                  <th className="text-left px-4 py-2.5 font-medium text-xs text-muted-foreground">地区</th>
                  <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground hidden sm:table-cell">独立 IP</th>
                  <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground">连接数</th>
                  <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground hidden sm:table-cell">上传</th>
                  <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground">下载</th>
                </tr>
              </thead>
              <tbody>
                {regions.length === 0 && (
                  <tr><td colSpan={6} className="text-center py-8 text-muted-foreground text-xs">暂无数据</td></tr>
                )}
                {regions.map((r, i) => (
                  <tr key={r.country_code} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-2.5 text-xs text-muted-foreground font-mono">{i + 1}</td>
                    <td className="px-4 py-2.5 text-sm">
                      <span className="mr-1.5">{r.flag}</span>
                      {r.country_name}
                    </td>
                    <td className="px-4 py-2.5 text-right text-xs text-muted-foreground font-mono hidden sm:table-cell">{r.ips}</td>
                    <td className="px-4 py-2.5 text-right text-xs font-mono">{r.count}</td>
                    <td className="px-4 py-2.5 text-right text-xs font-mono text-blue-500 hidden sm:table-cell">{formatBytes(r.upload)}</td>
                    <td className="px-4 py-2.5 text-right text-xs font-mono text-emerald-500">{formatBytes(r.download)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Top 来源 IP */}
      <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
        <div className="px-5 py-3.5 border-b">
          <h3 className="font-semibold text-sm">Top 来源 IP</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/40">
                <th className="text-left px-4 py-2.5 font-medium text-xs text-muted-foreground">#</th>
                <th className="text-left px-4 py-2.5 font-medium text-xs text-muted-foreground">来源 IP</th>
                <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground">连接数</th>
                <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground hidden sm:table-cell">上传</th>
                <th className="text-right px-4 py-2.5 font-medium text-xs text-muted-foreground">下载</th>
              </tr>
            </thead>
            <tbody>
              {sourceIPs.length === 0 && (
                <tr><td colSpan={5} className="text-center py-8 text-muted-foreground text-xs">暂无数据</td></tr>
              )}
              {sourceIPs.map((ip, i) => (
                <tr key={ip.source_ip} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-2.5 text-xs text-muted-foreground font-mono">{i + 1}</td>
                  <td className="px-4 py-2.5 font-mono text-xs">{ip.source_ip}</td>
                  <td className="px-4 py-2.5 text-right text-xs font-mono">{ip.count}</td>
                  <td className="px-4 py-2.5 text-right text-xs font-mono text-blue-500 hidden sm:table-cell">{formatBytes(ip.upload)}</td>
                  <td className="px-4 py-2.5 text-right text-xs font-mono text-emerald-500">{formatBytes(ip.download)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
