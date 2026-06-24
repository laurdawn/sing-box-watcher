import { useEffect, useState } from 'react'
import { api, GeoInfo } from '@/lib/api'
export type { GeoInfo }

export function useGeo(ips: string[]) {
  const [geoMap, setGeoMap] = useState<Record<string, GeoInfo>>({})

  useEffect(() => {
    const unique = [...new Set(ips.filter(Boolean))]
    const missing = unique.filter(ip => !geoMap[ip])
    if (missing.length === 0) return

    api.geoLookup(missing).then(result => {
      setGeoMap(prev => ({ ...prev, ...result }))
    }).catch(() => {})
  }, [ips.join(',')])

  return geoMap
}

// 国旗 emoji：ISO 3166-1 alpha-2 转区域指示符
export function countryFlag(code: string): string {
  if (!code || code.length !== 2) return ''
  return String.fromCodePoint(
    ...code.toUpperCase().split('').map(c => 0x1F1E6 + c.charCodeAt(0) - 65)
  )
}

// 生成可读的地理位置标签，如 "🇨🇳 中国 · 广东 · 深圳"
export function geoLabel(geo?: GeoInfo): string {
  if (!geo?.country_code) return ''
  const flag = countryFlag(geo.country_code)
  const parts = [flag + ' ' + geo.country_name]
  if (geo.region && geo.region !== geo.country_name) parts.push(geo.region)
  if (geo.city && geo.city !== geo.region) parts.push(geo.city)
  return parts.join(' · ')
}
