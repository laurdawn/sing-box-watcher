import { useState, useEffect, useRef, useCallback } from 'react'
import { GroupsSnapshot, api } from '@/lib/api'

export function useGroups(instance: string) {
  const [snapshot, setSnapshot] = useState<GroupsSnapshot | null>(null)
  const [testing, setTesting] = useState<Set<string>>(new Set())
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (!instance) return

    setSnapshot(null)
    let cancelled = false

    const connect = () => {
      if (cancelled) return
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const ws = new WebSocket(`${proto}://${window.location.host}/ws/groups?instance=${encodeURIComponent(instance)}`)
      wsRef.current = ws

      ws.onmessage = (e) => {
        try {
          setSnapshot(JSON.parse(e.data))
        } catch (_) {}
      }
      ws.onclose = () => { if (!cancelled) setTimeout(connect, 3000) }
    }

    connect()
    return () => {
      cancelled = true
      wsRef.current?.close()
    }
  }, [instance])

  const selectOutbound = useCallback(async (groupTag: string, outboundTag: string) => {
    await api.selectOutbound(instance, groupTag, outboundTag)
    const snap = await api.groups(instance)
    if (snap) setSnapshot(snap)
  }, [instance])

  const urlTest = useCallback(async (outboundTag: string) => {
    setTesting(prev => new Set(prev).add(outboundTag))
    try {
      const before = await api.groups(instance)
      const baseUpdatedAt = before?.updatedAt ?? 0
      await api.urlTest(instance, outboundTag)
      // 轮询直到 updatedAt 变化（数据刷新），最多等 10 秒
      for (let i = 0; i < 10; i++) {
        await new Promise(r => setTimeout(r, 1000))
        const snap = await api.groups(instance)
        if (snap) setSnapshot(snap)
        if (snap && snap.updatedAt !== baseUpdatedAt) break
      }
    } finally {
      setTesting(prev => {
        const next = new Set(prev)
        next.delete(outboundTag)
        return next
      })
    }
  }, [instance])

  return { snapshot, selectOutbound, urlTest, testing }
}
