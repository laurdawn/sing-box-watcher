import { useState, useEffect, useRef } from 'react'

export interface LiveTraffic {
  instance: string
  up: number
  down: number
  ts: number
}

export function useWsTraffic() {
  const [traffic, setTraffic] = useState<LiveTraffic[]>([])
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    const connect = () => {
      const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const ws = new WebSocket(`${proto}://${window.location.host}/ws/traffic`)
      wsRef.current = ws

      ws.onmessage = (e) => {
        try {
          const data: LiveTraffic[] = JSON.parse(e.data)
          setTraffic(data)
        } catch (_) {}
      }
      ws.onclose = () => {
        setTimeout(connect, 3000)
      }
    }
    connect()
    return () => wsRef.current?.close()
  }, [])

  return traffic
}
