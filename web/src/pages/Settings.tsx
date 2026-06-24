import { useEffect, useState } from 'react'
import { Plus, Trash2, Save, RotateCcw } from 'lucide-react'

interface Instance {
  name: string
  api: string
  secret: string
}

interface ConfigData {
  retention_days: number
  geo_db_path: string
  geo_db_url: string
  instances: Instance[]
}

const emptyInstance = (): Instance => ({ name: '', api: '', secret: '' })
const defaultCfg = (): ConfigData => ({ retention_days: 7, geo_db_path: '', geo_db_url: '', instances: [] })

export function Settings() {
  const [cfg, setCfg] = useState<ConfigData>(defaultCfg())
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState<{ type: 'ok' | 'err'; text: string } | null>(null)

  useEffect(() => {
    fetch('/api/config')
      .then(r => r.json())
      .then((data: ConfigData) => setCfg({ ...defaultCfg(), ...data, instances: data.instances ?? [] }))
      .catch(() => {})
  }, [])

  const setInstance = (i: number, key: keyof Instance, val: string) => {
    setCfg(prev => {
      const instances = [...prev.instances]
      instances[i] = { ...instances[i], [key]: val }
      return { ...prev, instances }
    })
  }

  const addInstance = () => setCfg(prev => ({ ...prev, instances: [...prev.instances, emptyInstance()] }))
  const removeInstance = (i: number) =>
    setCfg(prev => ({ ...prev, instances: prev.instances.filter((_, idx) => idx !== i) }))

  const save = async () => {
    setSaving(true)
    setMsg(null)
    try {
      const res = await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cfg),
      })
      if (!res.ok) {
        const text = await res.text()
        setMsg({ type: 'err', text })
      } else {
        setMsg({ type: 'ok', text: '保存成功，采集器已热重载' })
      }
    } catch (e) {
      setMsg({ type: 'err', text: String(e) })
    } finally {
      setSaving(false)
    }
  }

  const reset = () => {
    setMsg(null)
    fetch('/api/config').then(r => r.json()).then((data: ConfigData) => setCfg({ ...defaultCfg(), ...data, instances: data.instances ?? [] }))
  }

  return (
    <div className="max-w-2xl space-y-8">
      {/* 基础设置 */}
      <div>
        <h2 className="text-base font-semibold mb-1">基础设置</h2>
        <p className="text-sm text-muted-foreground mb-4">修改后点击保存，采集器自动热重载，无需重启。</p>
        <div className="rounded-xl border bg-card p-5 space-y-4">
          <div className="flex items-center gap-4">
            <label className="text-sm font-medium w-28 shrink-0">数据保留天数</label>
            <input
              type="number"
              min={1}
              max={365}
              value={cfg.retention_days}
              onChange={e => setCfg(prev => ({ ...prev, retention_days: Number(e.target.value) }))}
              className="h-9 w-28 rounded-md border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
            />
            <span className="text-sm text-muted-foreground">天（默认 7 天）</span>
          </div>
        </div>
      </div>

      {/* GeoIP 设置 */}
      <div>
        <h2 className="text-base font-semibold mb-1">GeoIP 数据库</h2>
        <p className="text-sm text-muted-foreground mb-4">用于 IP 归属地查询。首次启动若文件不存在会自动下载。</p>
        <div className="rounded-xl border bg-card p-5 space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">数据库路径</label>
            <input
              value={cfg.geo_db_path}
              onChange={e => setCfg(prev => ({ ...prev, geo_db_path: e.target.value }))}
              placeholder="./data/GeoLite2-City.mmdb"
              className="h-9 w-full rounded-md border bg-background px-3 text-sm font-mono focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">
              下载地址 <span className="font-normal">（留空使用默认源）</span>
            </label>
            <input
              value={cfg.geo_db_url}
              onChange={e => setCfg(prev => ({ ...prev, geo_db_url: e.target.value }))}
              placeholder="https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"
              className="h-9 w-full rounded-md border bg-background px-3 text-sm font-mono focus:outline-none focus:ring-1 focus:ring-primary"
            />
          </div>
        </div>
      </div>

      {/* sing-box 实例 */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-base font-semibold">sing-box 实例</h2>
            <p className="text-sm text-muted-foreground mt-0.5">每个实例独立采集流量和连接数据。</p>
          </div>
          <button
            onClick={addInstance}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md border text-sm hover:bg-accent transition-colors"
          >
            <Plus className="w-3.5 h-3.5" /> 添加实例
          </button>
        </div>
        <div className="space-y-3">
          {cfg.instances.length === 0 && (
            <div className="rounded-xl border border-dashed p-8 text-center text-muted-foreground text-sm">
              还没有实例，点击右上角添加
            </div>
          )}
          {cfg.instances.map((inst, i) => (
            <div key={i} className="rounded-xl border bg-card p-5 space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-muted-foreground">实例 {i + 1}</span>
                <button
                  onClick={() => removeInstance(i)}
                  className="p-1.5 rounded-md text-muted-foreground hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">名称</label>
                  <input
                    value={inst.name}
                    onChange={e => setInstance(i, 'name', e.target.value)}
                    placeholder="vps-hk"
                    className="h-9 w-full rounded-md border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">API 地址</label>
                  <input
                    value={inst.api}
                    onChange={e => setInstance(i, 'api', e.target.value)}
                    placeholder="https://your-vps:19090"
                    className="h-9 w-full rounded-md border bg-background px-3 text-sm font-mono focus:outline-none focus:ring-1 focus:ring-primary"
                  />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">
                  Secret <span className="font-normal">（对应 sing-box api.secret，留空则不认证）</span>
                </label>
                <input
                  type="password"
                  value={inst.secret}
                  onChange={e => setInstance(i, 'secret', e.target.value)}
                  placeholder="留空则不认证"
                  className="h-9 w-64 rounded-md border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="flex items-center gap-3 pt-2">
        <button
          onClick={save}
          disabled={saving}
          className="flex items-center gap-2 px-4 py-2 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:opacity-90 disabled:opacity-50 transition-opacity"
        >
          <Save className="w-4 h-4" />
          {saving ? '保存中...' : '保存'}
        </button>
        <button
          onClick={reset}
          className="flex items-center gap-2 px-4 py-2 rounded-md border text-sm text-muted-foreground hover:bg-accent transition-colors"
        >
          <RotateCcw className="w-4 h-4" /> 重置
        </button>
        {msg && (
          <span className={`text-sm ${msg.type === 'ok' ? 'text-emerald-600' : 'text-red-500'}`}>
            {msg.text}
          </span>
        )}
      </div>
    </div>
  )
}
