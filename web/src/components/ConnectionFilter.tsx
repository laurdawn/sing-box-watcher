import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { Search, X, SlidersHorizontal } from 'lucide-react'

export interface FilterState {
  inbound: string
  outbound: string
  source: string
  search: string
  rule: string
  from: string
  to: string
}

interface Props {
  instance: string
  value: FilterState
  onChange: (f: FilterState) => void
}

export function ConnectionFilter({ instance, value, onChange }: Props) {
  const [inbounds, setInbounds] = useState<string[]>([])
  const [outbounds, setOutbounds] = useState<string[]>([])

  useEffect(() => {
    if (!instance) return
    api.inbounds(instance).then(setInbounds).catch(() => setInbounds([]))
    api.outbounds(instance).then(setOutbounds).catch(() => setOutbounds([]))
  }, [instance])

  const set = (key: keyof FilterState, v: string) => onChange({ ...value, [key]: v })
  const clear = () => onChange({ inbound: '', outbound: '', source: '', search: '', rule: '', from: '', to: '' })
  const hasFilter = Object.values(value).some(v => v !== '')

  const selectCls = 'h-8 rounded-lg border bg-background px-3 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 transition-colors'
  const inputCls = 'h-8 rounded-lg border bg-background text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 transition-colors'

  return (
    <div className="rounded-xl border bg-card shadow-sm p-3">
      <div className="flex items-center gap-2 mb-2.5">
        <SlidersHorizontal className="w-3.5 h-3.5 text-muted-foreground" />
        <span className="text-xs font-medium text-muted-foreground">筛选</span>
        {hasFilter && (
          <button onClick={clear} className="ml-auto flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors">
            <X className="w-3 h-3" /> 清除
          </button>
        )}
      </div>
      <div className="flex flex-wrap gap-2">
        <select value={value.inbound} onChange={e => set('inbound', e.target.value)} className={selectCls}>
          <option value="">全部入站</option>
          {inbounds.map(i => <option key={i} value={i}>{i}</option>)}
        </select>

        <select value={value.outbound} onChange={e => set('outbound', e.target.value)} className={selectCls}>
          <option value="">全部出站</option>
          {outbounds.map(o => <option key={o} value={o}>{o}</option>)}
        </select>

        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
          <input value={value.source} onChange={e => set('source', e.target.value)} placeholder="源 IP"
            className={`${inputCls} pl-7 pr-3 w-32`} />
        </div>

        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
          <input value={value.search} onChange={e => set('search', e.target.value)} placeholder="目标域名 / IP"
            className={`${inputCls} pl-7 pr-3 w-40`} />
        </div>

        <input value={value.rule} onChange={e => set('rule', e.target.value)} placeholder="规则"
          className={`${inputCls} px-3 w-32`} />

        <input type="datetime-local" value={value.from} onChange={e => set('from', e.target.value)}
          className={`${inputCls} px-3`} />
        <span className="text-muted-foreground text-xs self-center">—</span>
        <input type="datetime-local" value={value.to} onChange={e => set('to', e.target.value)}
          className={`${inputCls} px-3`} />
      </div>
    </div>
  )
}
