import { useEffect, useRef, useState } from 'react'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, ReferenceArea,
} from 'recharts'
import { api, TrafficPoint } from '@/lib/api'
import { formatBytes } from '@/lib/utils'
import { cn } from '@/lib/utils'

interface Props {
  instance: string
  onRangeSelect?: (from: number, to: number) => void
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

export function TrafficChart({ instance, onRangeSelect }: Props) {
  const [range, setRange] = useState<Range>('1h')
  const [points, setPoints] = useState<TrafficPoint[]>([])

  const [dragStart, setDragStart] = useState<number | null>(null)
  const [dragEnd, setDragEnd] = useState<number | null>(null)
  const [selection, setSelection] = useState<{ from: number; to: number } | null>(null)
  const isDragging = useRef(false)

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

  useEffect(() => {
    setDragStart(null)
    setDragEnd(null)
    setSelection(null)
  }, [range, instance])

  const handleMouseDown = (e: { activeLabel?: number | string }) => {
    if (e?.activeLabel == null) return
    const ts = Number(e.activeLabel)
    isDragging.current = true
    setDragStart(ts)
    setDragEnd(ts)
    setSelection(null)
  }

  const handleMouseMove = (e: { activeLabel?: number | string }) => {
    if (!isDragging.current || e?.activeLabel == null) return
    setDragEnd(Number(e.activeLabel))
  }

  const handleMouseUp = () => {
    if (!isDragging.current) return
    isDragging.current = false
    if (dragStart != null && dragEnd != null && dragStart !== dragEnd) {
      const from = Math.min(dragStart, dragEnd)
      const to = Math.max(dragStart, dragEnd)
      setSelection({ from, to })
      onRangeSelect?.(from, to)
    }
    setDragStart(null)
    setDragEnd(null)
  }

  const selFrom = dragStart != null && dragEnd != null
    ? Math.min(dragStart, dragEnd)
    : selection?.from

  const selTo = dragStart != null && dragEnd != null
    ? Math.max(dragStart, dragEnd)
    : selection?.to

  return (
    <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
      <div className="flex items-center justify-between px-5 py-4 border-b">
        <div className="flex items-center gap-3">
          <h3 className="font-semibold text-sm">流量历史</h3>
          {selection && onRangeSelect && (
            <span className="text-xs text-muted-foreground">
              已选区间 ·{' '}
              <button onClick={() => setSelection(null)} className="text-blue-500 hover:underline">
                清除
              </button>
            </span>
          )}
        </div>
        <div className="flex gap-1">
          {RANGES.map(r => (
            <button
              key={r.value}
              onClick={() => setRange(r.value)}
              className={cn(
                'px-3 py-1 text-xs rounded-md transition-colors font-medium',
                range === r.value
                  ? 'bg-blue-600 text-white'
                  : 'text-muted-foreground hover:bg-accent'
              )}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>
      {onRangeSelect && (
        <p className="text-xs text-muted-foreground px-5 pt-3 -mb-1">拖拽图表选择时间区间，自动筛选连接</p>
      )}
      <div className="px-2 py-4">
        <ResponsiveContainer width="100%" height={220}>
          <AreaChart
            data={points}
            margin={{ top: 4, right: 8, left: 0, bottom: 0 }}
            onMouseDown={handleMouseDown}
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
            style={{ userSelect: 'none' }}
          >
            <defs>
              <linearGradient id="colorUp" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.25} />
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="colorDown" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#10b981" stopOpacity={0.25} />
                <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" vertical={false} />
            <XAxis
              dataKey="ts"
              tickFormatter={ts => formatXTick(ts, range)}
              tick={{ fontSize: 10, fill: 'hsl(var(--muted-foreground))' }}
              axisLine={false}
              tickLine={false}
              tickMargin={8}
            />
            <YAxis
              tickFormatter={v => formatBytes(v)}
              tick={{ fontSize: 10, fill: 'hsl(var(--muted-foreground))' }}
              axisLine={false}
              tickLine={false}
              width={68}
            />
            <Tooltip
              formatter={(v: number, name: string) => [formatBytes(v) + '/s', name === 'upload' ? '上传' : '下载']}
              labelFormatter={ts => formatXTick(Number(ts), range)}
              contentStyle={{
                background: 'hsl(var(--card))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '8px',
                fontSize: 12,
                boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)',
              }}
            />
            <Area type="monotone" dataKey="upload" stroke="#3b82f6" strokeWidth={1.5} fill="url(#colorUp)" name="upload" dot={false} />
            <Area type="monotone" dataKey="download" stroke="#10b981" strokeWidth={1.5} fill="url(#colorDown)" name="download" dot={false} />
            {selFrom != null && selTo != null && selFrom !== selTo && (
              <ReferenceArea x1={selFrom} x2={selTo} strokeOpacity={0.3} fill="#3b82f6" fillOpacity={0.12} />
            )}
          </AreaChart>
        </ResponsiveContainer>
        <div className="flex items-center gap-4 px-3 mt-1">
          <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span className="w-3 h-0.5 bg-blue-500 rounded inline-block" />上传
          </span>
          <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span className="w-3 h-0.5 bg-emerald-500 rounded inline-block" />下载
          </span>
        </div>
      </div>
    </div>
  )
}
