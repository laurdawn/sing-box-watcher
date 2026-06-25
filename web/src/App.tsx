import { useState, useEffect } from 'react'
import { Activity, BarChart2, List, Box, Settings2, Radio, ScrollText, LogOut, ChevronDown } from 'lucide-react'
import { useInstances } from '@/hooks/useInstances'
import { Dashboard } from '@/pages/Dashboard'
import { Connections } from '@/pages/Connections'
import { Analysis } from '@/pages/Analysis'
import { Settings } from '@/pages/Settings'
import { Proxies } from '@/pages/Proxies'
import { Logs } from '@/pages/Logs'
import { Login } from '@/pages/Login'
import { cn } from '@/lib/utils'

type Page = 'dashboard' | 'connections' | 'analysis' | 'proxies' | 'logs' | 'settings'

const NAV: { id: Page; label: string; icon: React.ReactNode }[] = [
  { id: 'dashboard', label: '总览', icon: <Activity className="w-4 h-4" /> },
  { id: 'proxies', label: '代理节点', icon: <Radio className="w-4 h-4" /> },
  { id: 'connections', label: '连接历史', icon: <List className="w-4 h-4" /> },
  { id: 'analysis', label: '流量分析', icon: <BarChart2 className="w-4 h-4" /> },
  { id: 'logs', label: '日志', icon: <ScrollText className="w-4 h-4" /> },
  { id: 'settings', label: '设置', icon: <Settings2 className="w-4 h-4" /> },
]

export default function App() {
  const [page, setPage] = useState<Page>('dashboard')
  const [authed, setAuthed] = useState<boolean | null>(null)
  const { instances, selected, setSelected } = useInstances()

  useEffect(() => {
    fetch('/api/auth/me')
      .then(r => setAuthed(r.ok))
      .catch(() => setAuthed(false))
  }, [])

  const logout = async () => {
    await fetch('/api/auth/logout', { method: 'POST' })
    setAuthed(false)
  }

  if (authed === null) return null

  if (!authed) return <Login onSuccess={() => setAuthed(true)} />

  const current = instances.find(i => i.name === selected)

  return (
    <div className="min-h-screen flex bg-background">
      {/* 侧边栏 — 桌面端 */}
      <aside className="hidden md:flex flex-col w-52 shrink-0 bg-slate-900 dark:bg-slate-950 min-h-screen sticky top-0 h-screen">
        {/* logo */}
        <div className="flex items-center gap-2.5 px-5 h-14 border-b border-slate-700/50">
          <div className="w-7 h-7 rounded-lg bg-blue-500 flex items-center justify-center">
            <Box className="w-4 h-4 text-white" />
          </div>
          <span className="font-semibold text-sm text-white tracking-tight">sing-box watcher</span>
        </div>

        {/* 实例选择 */}
        {instances.length > 0 && page !== 'settings' && (
          <div className="px-3 pt-4 pb-2">
            <div className="relative">
              <select
                value={selected}
                onChange={e => setSelected(e.target.value)}
                className="w-full appearance-none bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs rounded-lg px-3 py-2 pl-7 pr-7 focus:outline-none focus:ring-1 focus:ring-blue-500 cursor-pointer border border-slate-700/50 transition-colors"
              >
                {instances.map(inst => (
                  <option key={inst.name} value={inst.name}>{inst.name}</option>
                ))}
              </select>
              <span className={cn(
                'pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 w-1.5 h-1.5 rounded-full',
                current?.online ? 'bg-emerald-400' : 'bg-red-400'
              )} />
              <ChevronDown className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400" />
            </div>
          </div>
        )}

        {/* 导航 */}
        <nav className="flex-1 px-3 py-3 space-y-0.5">
          {NAV.map(item => (
            <button
              key={item.id}
              onClick={() => setPage(item.id)}
              className={cn(
                'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all',
                page === item.id
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
              )}
            >
              {item.icon}
              {item.label}
            </button>
          ))}
        </nav>

        {/* 退出 */}
        <div className="px-3 py-4 border-t border-slate-700/50">
          <button
            onClick={logout}
            className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition-colors"
          >
            <LogOut className="w-4 h-4" />
            退出登录
          </button>
        </div>
      </aside>

      {/* 内容区 */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* 移动端顶部栏 */}
        <header className="md:hidden sticky top-0 z-10 bg-slate-900 border-b border-slate-700/50">
          <div className="h-14 px-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-6 h-6 rounded-md bg-blue-500 flex items-center justify-center">
                <Box className="w-3.5 h-3.5 text-white" />
              </div>
              <span className="font-semibold text-sm text-white">sbw</span>
            </div>
            {instances.length > 0 && page !== 'settings' && (
              <div className="relative">
                <select
                  value={selected}
                  onChange={e => setSelected(e.target.value)}
                  className="appearance-none bg-slate-800 text-slate-200 text-xs rounded-lg px-3 py-1.5 pl-6 pr-6 focus:outline-none focus:ring-1 focus:ring-blue-500 border border-slate-700/50 max-w-[140px]"
                >
                  {instances.map(inst => (
                    <option key={inst.name} value={inst.name}>{inst.name}</option>
                  ))}
                </select>
                <span className={cn(
                  'pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 w-1.5 h-1.5 rounded-full',
                  current?.online ? 'bg-emerald-400' : 'bg-red-400'
                )} />
                <ChevronDown className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 w-3 h-3 text-slate-400" />
              </div>
            )}
          </div>
        </header>

        <main className="flex-1 p-4 sm:p-6 min-w-0 pb-20 md:pb-8">
          {!selected && page !== 'settings' && page !== 'logs' && (
            <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
              正在连接 sing-box 实例...
            </div>
          )}
          {selected && page === 'dashboard' && <Dashboard instance={selected} instances={instances} />}
          {selected && page === 'proxies' && <Proxies instance={selected} />}
          {selected && page === 'connections' && <Connections instance={selected} />}
          {selected && page === 'analysis' && <Analysis instance={selected} />}
          {selected && page === 'logs' && <Logs instance={selected} />}
          {page === 'settings' && <Settings />}
        </main>
      </div>

      {/* 移动端底部导航 */}
      <nav className="md:hidden fixed bottom-0 left-0 right-0 z-10 bg-slate-900 border-t border-slate-700/50">
        <div className="flex items-stretch h-14">
          {NAV.map(item => (
            <button
              key={item.id}
              onClick={() => setPage(item.id)}
              className={cn(
                'flex-1 flex flex-col items-center justify-center gap-0.5 transition-colors',
                page === item.id ? 'text-blue-400' : 'text-slate-500'
              )}
            >
              <span className={cn('w-5 h-5', page === item.id && '[&>svg]:stroke-[2.5]')}>
                {item.icon}
              </span>
              <span className="text-[9px] font-medium">{item.label}</span>
            </button>
          ))}
        </div>
      </nav>
    </div>
  )
}
