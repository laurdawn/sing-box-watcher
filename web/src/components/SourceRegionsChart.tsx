import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Cell } from 'recharts'
import { RegionStat } from '@/lib/api'
import { formatBytes } from '@/lib/utils'

const COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#84cc16', '#f97316', '#ec4899', '#14b8a6']

interface Props {
  data: RegionStat[]
}

export function SourceRegionsChart({ data }: Props) {
  const top = data.slice(0, 12)

  return (
    <div className="rounded-xl border bg-card p-5 shadow-sm">
      <h3 className="font-medium mb-4">来源地区分布（连接次数）</h3>
      <ResponsiveContainer width="100%" height={320}>
        <BarChart data={top} layout="vertical" margin={{ left: 8, right: 24, top: 4, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" horizontal={false} />
          <XAxis
            type="number"
            tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
            axisLine={false}
            tickLine={false}
          />
          <YAxis
            type="category"
            dataKey="country_name"
            width={90}
            tick={({ x, y, payload, index }) => (
              <text x={x} y={y} textAnchor="end" dominantBaseline="middle" fontSize={12} fill="hsl(var(--foreground))">
                {top[index]?.flag} {payload.value}
              </text>
            )}
            axisLine={false}
            tickLine={false}
          />
          <Tooltip
            formatter={(v: number, name: string) => [
              name === 'count' ? `${v} 次` : formatBytes(v),
              name === 'count' ? '连接数' : name === 'upload' ? '上传' : '下载',
            ]}
            labelFormatter={(label) => {
              const item = top.find(d => d.country_name === label)
              return item ? `${item.flag} ${label}（${item.ips} 个 IP）` : label
            }}
            contentStyle={{
              background: 'hsl(var(--card))',
              border: '1px solid hsl(var(--border))',
              borderRadius: '8px',
              fontSize: 12,
            }}
          />
          <Bar dataKey="count" radius={[0, 4, 4, 0]}>
            {top.map((_, i) => (
              <Cell key={i} fill={COLORS[i % COLORS.length]} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
