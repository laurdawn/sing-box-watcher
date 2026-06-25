import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { TopOutbound } from '@/lib/api'
import { formatBytes } from '@/lib/utils'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#84cc16', '#f97316']

interface Props {
  data: TopOutbound[]
}

export function OutboundChart({ data }: Props) {
  const top = data.slice(0, 8)
  return (
    <div className="rounded-xl border bg-card p-5 shadow-sm">
      <h3 className="font-medium mb-4">出站流量分布</h3>
      <ResponsiveContainer width="100%" height={280}>
        <PieChart>
          <Pie
            data={top}
            dataKey="download"
            nameKey="outbound"
            cx="50%"
            cy="50%"
            innerRadius={60}
            outerRadius={100}
            paddingAngle={3}
          >
            {top.map((_, i) => (
              <Cell key={i} fill={COLORS[i % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip
            formatter={(v: number) => [formatBytes(v), '下载']}
            contentStyle={{
              background: 'hsl(var(--card))',
              border: '1px solid hsl(var(--border))',
              borderRadius: '8px',
              fontSize: 12,
            }}
          />
          <Legend
            formatter={(v) => v}
            iconType="circle"
            iconSize={8}
            wrapperStyle={{ fontSize: 12 }}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  )
}
