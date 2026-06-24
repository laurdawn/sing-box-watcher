import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { Search, X } from 'lucide-react'

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

  return (
    <div className="flex flex-wrap gap-3 items-center p-4 rounded-xl border bg-card">
      <select
        value={value.inbound}
        onChange={e => set('inbound', e.target.value)}
        className="h-9 rounded-md border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
      >
        <option value="">全部入站</option>
        {inbounds.map(i => <option key={i} value={i}>{i}</option>)}
      </select>

      <select
        value={value.outbound}
        onChange={e => set('outbound', e.target.value)}
        className="h-9 rounded-md border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
      >
        <option value="">全部出站</option>
        {outbounds.map(o => <option key={o} value={o}>{o}</option>)}
      </select>

      <div className="relative">
        <Search className="absolute left-2.5 top-2 w-4 h-4 text-muted-foreground" />
        <input
          value={value.source}
          onChange={e => set('source', e.target.value)}
          placeholder="源 IP"
          className="h-9 pl-8 pr-3 rounded-md border bg-background text-sm focus:outline-none focus:ring-1 focus:ring-primary w-36"
        />
      </div>

      <div className="relative">
        <Search className="absolute left-2.5 top-2 w-4 h-4 text-muted-foreground" />
        <input
          value={value.search}
          onChange={e => set('search', e.target.value)}
          placeholder="目标域名 / IP"
          className="h-9 pl-8 pr-3 rounded-md border bg-background text-sm focus:outline-none focus:ring-1 focus:ring-primary w-44"
        />
      </div>

      <input
        value={value.rule}
        onChange={e => set('rule', e.target.value)}
        placeholder="规则"
        className="h-9 px-3 rounded-md border bg-background text-sm focus:outline-none focus:ring-1 focus:ring-primary w-36"
      />

      <input
        type="datetime-local"
        value={value.from}
        onChange={e => set('from', e.target.value)}
        className="h-9 px-3 rounded-md border bg-background text-sm focus:outline-none focus:ring-1 focus:ring-primary"
      />
      <span className="text-muted-foreground text-sm">~</span>
      <input
        type="datetime-local"
        value={value.to}
        onChange={e => set('to', e.target.value)}
        className="h-9 px-3 rounded-md border bg-background text-sm focus:outline-none focus:ring-1 focus:ring-primary"
      />

      {hasFilter && (
        <button
          onClick={clear}
          className="flex items-center gap-1 h-9 px-3 rounded-md text-sm text-muted-foreground hover:bg-accent transition-colors"
        >
          <X className="w-3.5 h-3.5" /> 清除
        </button>
      )}
    </div>
  )
}
