import { useEffect, useState } from 'react'
import { Plus, Trash2, Save, RotateCcw, Copy, RefreshCw } from 'lucide-react'

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
  mcp_enabled: boolean
  log_persist_enabled: boolean
  log_persist_min_level: string
}

const emptyInstance = (): Instance => ({ name: '', api: '', secret: '' })
const defaultCfg = (): ConfigData => ({ retention_days: 7, geo_db_path: '', geo_db_url: '', instances: [], mcp_enabled: false, log_persist_enabled: false, log_persist_min_level: 'WARN' })

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

  const inputCls = 'h-9 rounded-lg border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 transition-colors'
  const sectionCls = 'rounded-xl border bg-card shadow-sm p-5 space-y-4'

  return (
    <div className="max-w-2xl space-y-7">
      {/* 基础设置 */}
      <section>
        <div className="mb-3">
          <h2 className="text-sm font-semibold">基础设置</h2>
          <p className="text-xs text-muted-foreground mt-0.5">修改后点击保存，采集器自动热重载，无需重启。</p>
        </div>
        <div className={sectionCls}>
          <div className="flex items-center gap-4">
            <label className="text-sm font-medium w-28 shrink-0">数据保留天数</label>
            <input
              type="number" min={1} max={365}
              value={cfg.retention_days}
              onChange={e => setCfg(prev => ({ ...prev, retention_days: Number(e.target.value) }))}
              className={`${inputCls} w-24`}
            />
            <span className="text-xs text-muted-foreground">天（默认 7 天）</span>
          </div>
        </div>
      </section>

      {/* GeoIP */}
      <section>
        <div className="mb-3">
          <h2 className="text-sm font-semibold">GeoIP 数据库</h2>
          <p className="text-xs text-muted-foreground mt-0.5">用于 IP 归属地查询。首次启动若文件不存在会自动下载。</p>
        </div>
        <div className={sectionCls}>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">数据库路径</label>
            <input value={cfg.geo_db_path} onChange={e => setCfg(prev => ({ ...prev, geo_db_path: e.target.value }))}
              placeholder="./data/GeoLite2-City.mmdb" className={`${inputCls} w-full font-mono text-xs`} />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">下载地址 <span className="font-normal opacity-60">（留空使用默认源）</span></label>
            <input value={cfg.geo_db_url} onChange={e => setCfg(prev => ({ ...prev, geo_db_url: e.target.value }))}
              placeholder="https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"
              className={`${inputCls} w-full font-mono text-xs`} />
          </div>
        </div>
      </section>

      {/* 实例 */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <div>
            <h2 className="text-sm font-semibold">sing-box 实例</h2>
            <p className="text-xs text-muted-foreground mt-0.5">每个实例独立采集流量和连接数据。</p>
          </div>
          <button onClick={addInstance}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium hover:bg-accent transition-colors">
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
            <div key={i} className="rounded-xl border bg-card shadow-sm p-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">实例 {i + 1}</span>
                <button onClick={() => removeInstance(i)}
                  className="p-1.5 rounded-lg text-muted-foreground hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">名称</label>
                  <input value={inst.name} onChange={e => setInstance(i, 'name', e.target.value)}
                    placeholder="vps-hk" className={`${inputCls} w-full`} />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">API 地址</label>
                  <input value={inst.api} onChange={e => setInstance(i, 'api', e.target.value)}
                    placeholder="https://your-vps:19090" className={`${inputCls} w-full font-mono text-xs`} />
                </div>
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">Secret <span className="font-normal opacity-60">（留空则不认证）</span></label>
                <input type="password" value={inst.secret} onChange={e => setInstance(i, 'secret', e.target.value)}
                  placeholder="留空则不认证" className={`${inputCls} w-64`} />
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* 日志持久化 */}
      <section>
        <div className="mb-3">
          <h2 className="text-sm font-semibold">日志持久化</h2>
          <p className="text-xs text-muted-foreground mt-0.5">开启后将日志写入 SQLite，支持历史查询。建议仅持久化 WARN 以上级别。</p>
        </div>
        <div className={sectionCls}>
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">启用持久化</p>
            <button role="switch" aria-checked={cfg.log_persist_enabled}
              onClick={() => setCfg(prev => ({ ...prev, log_persist_enabled: !prev.log_persist_enabled }))}
              className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${cfg.log_persist_enabled ? 'bg-blue-600' : 'bg-muted'}`}>
              <span className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform ${cfg.log_persist_enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
            </button>
          </div>
          {cfg.log_persist_enabled && (
            <div className="flex items-center gap-4">
              <label className="text-sm font-medium w-28 shrink-0">最低持久化级别</label>
              <select value={cfg.log_persist_min_level}
                onChange={e => setCfg(prev => ({ ...prev, log_persist_min_level: e.target.value }))}
                className={`${inputCls}`}>
                {['PANIC', 'FATAL', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'].map(l => (
                  <option key={l} value={l}>{l}</option>
                ))}
              </select>
            </div>
          )}
        </div>
      </section>

      {/* AI / MCP */}
      <section>
        <div className="mb-3">
          <h2 className="text-sm font-semibold">AI / MCP</h2>
          <p className="text-xs text-muted-foreground mt-0.5">启用后可通过 MCP 协议让 AI 直接分析流量和连接数据。</p>
        </div>
        <div className={sectionCls}>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">MCP Server</p>
              <p className="text-xs text-muted-foreground mt-0.5 font-mono">{window.location.origin}/mcp</p>
            </div>
            <button role="switch" aria-checked={cfg.mcp_enabled}
              onClick={() => setCfg(prev => ({ ...prev, mcp_enabled: !prev.mcp_enabled }))}
              className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${cfg.mcp_enabled ? 'bg-blue-600' : 'bg-muted'}`}>
              <span className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform ${cfg.mcp_enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
            </button>
          </div>
          <MCPTokenSection />
        </div>
      </section>

      {/* 修改密码 */}
      <ChangePasswordSection />

      {/* 操作按钮 */}
      <div className="flex items-center gap-3 pt-1">
        <button onClick={save} disabled={saving}
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium disabled:opacity-50 transition-colors">
          <Save className="w-4 h-4" />
          {saving ? '保存中...' : '保存'}
        </button>
        <button onClick={reset}
          className="flex items-center gap-2 px-4 py-2 rounded-lg border text-sm text-muted-foreground hover:bg-accent transition-colors">
          <RotateCcw className="w-4 h-4" /> 重置
        </button>
        {msg && (
          <span className={`text-xs ${msg.type === 'ok' ? 'text-emerald-500' : 'text-red-500'}`}>
            {msg.text}
          </span>
        )}
      </div>
    </div>
  )
}

function MCPTokenSection() {
  const [token, setToken] = useState('')
  const [copied, setCopied] = useState(false)
  const [regenerating, setRegenerating] = useState(false)

  useEffect(() => {
    fetch('/api/config').then(r => r.json()).then(d => setToken(d.mcp_token ?? ''))
  }, [])

  const copy = () => {
    const succeed = () => { setCopied(true); setTimeout(() => setCopied(false), 2000) }
    if (navigator.clipboard) {
      navigator.clipboard.writeText(token).then(succeed).catch(fallbackCopy)
    } else {
      fallbackCopy()
    }
  }

  const fallbackCopy = () => {
    const el = document.createElement('textarea')
    el.value = token
    el.style.cssText = 'position:fixed;opacity:0'
    document.body.appendChild(el)
    el.select()
    document.execCommand('copy')
    document.body.removeChild(el)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const regenerate = async () => {
    setRegenerating(true)
    try {
      const res = await fetch('/api/auth/regenerate-mcp-token', { method: 'POST' })
      if (res.ok) {
        const data = await res.json()
        setToken(data.mcp_token)
      }
    } finally {
      setRegenerating(false)
    }
  }

  if (!token) return null

  return (
    <div className="space-y-1.5">
      <label className="text-xs font-medium text-muted-foreground">API Token（Bearer）</label>
      <div className="flex items-center gap-2">
        <input readOnly value={token}
          className="h-8 flex-1 rounded-lg border bg-muted px-3 text-xs font-mono focus:outline-none" />
        <button onClick={copy}
          className="p-1.5 rounded-lg border hover:bg-accent transition-colors" title="复制">
          <Copy className="w-3.5 h-3.5" />
        </button>
        <button onClick={regenerate} disabled={regenerating}
          className="p-1.5 rounded-lg border hover:bg-accent transition-colors disabled:opacity-50" title="重新生成">
          <RefreshCw className={`w-3.5 h-3.5 ${regenerating ? 'animate-spin' : ''}`} />
        </button>
      </div>
      {copied && <p className="text-xs text-emerald-500">已复制</p>}
    </div>
  )
}

function ChangePasswordSection() {
  const [oldPwd, setOldPwd] = useState('')
  const [newPwd, setNewPwd] = useState('')
  const [confirm, setConfirm] = useState('')
  const [msg, setMsg] = useState<{ type: 'ok' | 'err'; text: string } | null>(null)
  const [saving, setSaving] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setMsg(null)
    if (newPwd !== confirm) { setMsg({ type: 'err', text: '两次输入的密码不一致' }); return }
    if (newPwd.length < 4) { setMsg({ type: 'err', text: '新密码至少 4 位' }); return }
    setSaving(true)
    try {
      const res = await fetch('/api/auth/password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ old_password: oldPwd, new_password: newPwd }),
      })
      if (res.ok) {
        setMsg({ type: 'ok', text: '密码已修改' })
        setOldPwd(''); setNewPwd(''); setConfirm('')
      } else {
        const text = await res.text()
        setMsg({ type: 'err', text })
      }
    } catch (e) {
      setMsg({ type: 'err', text: String(e) })
    } finally {
      setSaving(false)
    }
  }

  const inputCls = 'h-9 rounded-lg border bg-background px-3 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 transition-colors'

  return (
    <section>
      <div className="mb-3">
        <h2 className="text-sm font-semibold">修改密码</h2>
        <p className="text-xs text-muted-foreground mt-0.5">默认账号 admin / admin，建议修改。</p>
      </div>
      <form onSubmit={submit} className="rounded-xl border bg-card shadow-sm p-5 space-y-4">
        {[
          { label: '当前密码', value: oldPwd, set: setOldPwd },
          { label: '新密码', value: newPwd, set: setNewPwd },
          { label: '确认新密码', value: confirm, set: setConfirm },
        ].map(({ label, value, set }) => (
          <div key={label} className="flex items-center gap-4">
            <label className="text-sm font-medium w-28 shrink-0">{label}</label>
            <input type="password" value={value} onChange={e => set(e.target.value)}
              className={`${inputCls} w-56`} />
          </div>
        ))}
        <div className="flex items-center gap-3">
          <button type="submit" disabled={saving}
            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium disabled:opacity-50 transition-colors">
            {saving ? '保存中...' : '修改密码'}
          </button>
          {msg && <span className={`text-xs ${msg.type === 'ok' ? 'text-emerald-500' : 'text-red-500'}`}>{msg.text}</span>}
        </div>
      </form>
    </section>
  )
}
