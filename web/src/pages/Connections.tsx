import { useEffect, useState } from 'react'
import { api, Connection } from '@/lib/api'
import { ConnectionTable } from '@/components/ConnectionTable'
import { ConnectionFilter, FilterState } from '@/components/ConnectionFilter'
import { TrafficChart } from '@/components/TrafficChart'

interface Props {
  instance: string
}

type SortField = 'upload' | 'download' | ''
type SortDir = 'asc' | 'desc'

function tsToDatetimeLocal(ts: number): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function Connections({ instance }: Props) {
  const [conns, setConns] = useState<Connection[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [filter, setFilter] = useState<FilterState>({
    inbound: '', outbound: '', source: '', search: '', rule: '', from: '', to: '',
  })
  const [sortBy, setSortBy] = useState<SortField>('')
  const [sortDir, setSortDir] = useState<SortDir>('desc')

  const limit = 20

  useEffect(() => {
    setPage(1)
  }, [filter, instance, sortBy, sortDir])

  useEffect(() => {
    if (!instance) return
    const f = filter

    const fromTs = f.from ? new Date(f.from).getTime() : undefined
    const toTs = f.to ? new Date(f.to).getTime() : undefined

    api.connections({
      instance,
      inbound: f.inbound || undefined,
      outbound: f.outbound || undefined,
      source: f.source || undefined,
      search: f.search || undefined,
      rule: f.rule || undefined,
      from: fromTs,
      to: toTs,
      page,
      limit,
      sort_by: sortBy || undefined,
      sort_dir: sortBy ? sortDir : undefined,
    }).then(r => {
      setConns(r.connections || [])
      setTotal(r.total || 0)
    }).catch(() => {})
  }, [instance, filter, page, sortBy, sortDir])

  const handleRangeSelect = (from: number, to: number) => {
    setFilter(f => ({ ...f, from: tsToDatetimeLocal(from), to: tsToDatetimeLocal(to) }))
  }

  const handleSortChange = (field: SortField, dir: SortDir) => {
    setSortBy(field)
    setSortDir(dir)
  }

  return (
    <div className="space-y-4">
      <TrafficChart instance={instance} onRangeSelect={handleRangeSelect} />
      <ConnectionFilter instance={instance} value={filter} onChange={f => { setFilter(f); setPage(1) }} />
      <div className="text-xs text-muted-foreground px-1">共 {total} 条记录</div>
      <ConnectionTable
        connections={conns}
        total={total}
        page={page}
        limit={limit}
        onPageChange={setPage}
        sortBy={sortBy}
        sortDir={sortDir}
        onSortChange={handleSortChange}
      />
    </div>
  )
}
