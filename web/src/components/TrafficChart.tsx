import { useEffect, useState } from 'react'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend,
} from 'recharts'
import { api, TrafficPoint } from '@/lib/api'
import { formatBytes } from '@/lib/utils'

interface Props {
  instance: string
}

type Range = '1h' | '6h' | '24h' | '7d'

const RANGES: { label: string; value: Range; ms: number }[] = [
  { label: '1小时', value: '1h', ms: 3_600_000 },
  { label: '6小时', value: '6h', ms: 21_600_000 },
  { label: '24小时', value: '24h', ms: 86_400_000 },
  { label: '7天', value: '7d', ms: 604_800_000 },
]

function formatXTick(ts: number, range: Range) {
  const d = new Date(ts)
  if (range === '7d') return `${d.getMonth() + 1}/${d.getDate()}`
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

export function TrafficChart({ instance }: Props) {
  const [range, setRange] = useState<Range>('1h')
  const [points, setPoints] = useState<TrafficPoint[]>([])

  useEffect(() => {
    const load = async () => {
      const r = RANGES.find(r => r.value === range)!
      const to = Date.now()
      const from = to - r.ms
      try {
        const res = await api.traffic(instance, from, to)
        setPoints(res.points || [])
      } catch (_) {}
    }
    load()
    const t = setInterval(load, 10000)
    return () => clearInterval(t)
  }, [instance, range])

  return (
    <div className="rounded-xl border bg-card p-5 shadow-sm">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-medium">流量历史</h3>
        <div className="flex gap-1">
          {RANGES.map(r => (
            <button
              key={r.value}
              onClick={() => setRange(r.value)}
              className={`px-3 py-1 text-xs rounded-md transition-colors ${
                range === r.value
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent'
              }`}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>
      <ResponsiveContainer width="100%" height={240}>
        <AreaChart data={points} margin={{ top: 4, right: 4, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="colorUp" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorDown" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
          <XAxis
            dataKey="ts"
            tickFormatter={ts => formatXTick(ts, range)}
            tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
            axisLine={false}
            tickLine={false}
          />
          <YAxis
            tickFormatter={v => formatBytes(v)}
            tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
            axisLine={false}
            tickLine={false}
            width={72}
          />
          <Tooltip
            formatter={(v: number, name: string) => [formatBytes(v) + '/s', name === 'upload' ? '上传' : '下载']}
            labelFormatter={ts => formatXTick(Number(ts), range)}
            contentStyle={{
              background: 'hsl(var(--card))',
              border: '1px solid hsl(var(--border))',
              borderRadius: '8px',
              fontSize: 12,
            }}
          />
          <Legend
            formatter={v => v === 'upload' ? '上传' : '下载'}
            iconType="circle"
            iconSize={8}
          />
          <Area type="monotone" dataKey="upload" stroke="#6366f1" strokeWidth={2} fill="url(#colorUp)" />
          <Area type="monotone" dataKey="download" stroke="#10b981" strokeWidth={2} fill="url(#colorDown)" />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}
