import { useEffect, useState } from 'react'
import { api, Connection } from '@/lib/api'
import { ConnectionTable } from '@/components/ConnectionTable'
import { ConnectionFilter, FilterState } from '@/components/ConnectionFilter'

interface Props {
  instance: string
}

export function Connections({ instance }: Props) {
  const [conns, setConns] = useState<Connection[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [filter, setFilter] = useState<FilterState>({
    inbound: '', outbound: '', source: '', search: '', rule: '', from: '', to: '',
  })

  const limit = 20

  useEffect(() => {
    setPage(1)
  }, [filter, instance])

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
    }).then(r => {
      setConns(r.connections || [])
      setTotal(r.total || 0)
    }).catch(() => {})
  }, [instance, filter, page])

  return (
    <div className="space-y-4">
      <ConnectionFilter instance={instance} value={filter} onChange={f => { setFilter(f); setPage(1) }} />
      <div className="text-xs text-muted-foreground px-1">共 {total} 条记录</div>
      <ConnectionTable
        connections={conns}
        total={total}
        page={page}
        limit={limit}
        onPageChange={setPage}
      />
    </div>
  )
}
