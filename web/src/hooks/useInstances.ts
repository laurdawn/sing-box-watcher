import { useState, useEffect, useRef } from 'react'
import { api, InstanceStats } from '@/lib/api'

export function useInstances() {
  const [instances, setInstances] = useState<InstanceStats[]>([])
  const [selected, setSelected] = useState<string>('')
  const initializedRef = useRef(false)

  useEffect(() => {
    const load = async () => {
      try {
        const data = await api.instances()
        setInstances(data || [])
        if (data?.length && !initializedRef.current) {
          initializedRef.current = true
          setSelected(data[0].name)
        }
      } catch (_) {}
    }
    load()
    const t = setInterval(load, 3000)
    return () => clearInterval(t)
  }, [])

  return { instances, selected, setSelected }
}
