import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import { TopDomain } from '@/lib/api'
import { formatBytes } from '@/lib/utils'

interface Props {
  data: TopDomain[]
}

export function TopDomainsChart({ data }: Props) {
  const top = data.slice(0, 10)
  return (
    <div className="rounded-xl border bg-card p-5 shadow-sm">
      <h3 className="font-medium mb-4">Top 域名（连接次数）</h3>
      <ResponsiveContainer width="100%" height={280}>
        <BarChart data={top} layout="vertical" margin={{ left: 8, right: 16 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" horizontal={false} />
          <XAxis type="number" tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }} axisLine={false} tickLine={false} />
          <YAxis
            type="category"
            dataKey="host"
            width={160}
            tick={{ fontSize: 11, fill: 'hsl(var(--foreground))' }}
            axisLine={false}
            tickLine={false}
          />
          <Tooltip
            formatter={(v: number, name: string) => [
              name === 'count' ? `${v} 次` : formatBytes(v),
              name === 'count' ? '连接数' : name === 'upload' ? '上传' : '下载',
            ]}
            contentStyle={{
              background: 'hsl(var(--card))',
              border: '1px solid hsl(var(--border))',
              borderRadius: '8px',
              fontSize: 12,
            }}
          />
          <Bar dataKey="count" fill="#6366f1" radius={[0, 4, 4, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
