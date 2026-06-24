import { useState, useEffect, useRef } from 'react'
import { LogEntry } from '@/lib/api'

const MAX_LOGS = 500

export function useLogs(instance: string) {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (!instance) return
    setLogs([])

    const connect = () => {
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const ws = new WebSocket(`${proto}://${window.location.host}/ws/log?instance=${encodeURIComponent(instance)}`)
      wsRef.current = ws

      ws.onmessage = (e) => {
        try {
          const entry = JSON.parse(e.data) as LogEntry & { reset?: boolean }
          if (entry.reset) {
            setLogs([])
            return
          }
          // 去掉消息里 sing-box 自带的级别前缀，如 "INFO[418318] " 或 "WARN[123] "
          entry.message = entry.message.replace(/^\x1b\[[0-9;]*m[A-Z]+\x1b\[[0-9;]*m\[\d+\]\s*/, '')
            .replace(/^[A-Z]+\[\d+\]\s*/, '')
          setLogs(prev => {
            const next = [...prev, entry]
            return next.length > MAX_LOGS ? next.slice(next.length - MAX_LOGS) : next
          })
        } catch (_) {}
      }
      ws.onclose = () => setTimeout(connect, 3000)
    }

    connect()
    return () => wsRef.current?.close()
  }, [instance])

  const clear = () => setLogs([])

  return { logs, clear }
}
