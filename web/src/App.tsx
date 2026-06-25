import { useState, useEffect } from 'react'
import { Activity, BarChart2, List, Box, Settings2, Radio, ScrollText, LogOut } from 'lucide-react'
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
  const [authed, setAuthed] = useState<boolean | null>(null) // null = checking
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

  if (authed === null) return null // loading

  if (!authed) return <Login onSuccess={() => setAuthed(true)} />

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <header className="border-b sticky top-0 z-10 bg-background/80 backdrop-blur-sm">
        <div className="max-w-screen-xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between">
          <div className="flex items-center gap-2 font-semibold text-sm">
            <Box className="w-5 h-5 text-indigo-500" />
            <span className="hidden sm:inline">sing-box watcher</span>
            <span className="sm:hidden">sbw</span>
          </div>
          <div className="flex items-center gap-2">
            {instances.length > 0 && page !== 'settings' && (
              <div className="relative">
                {(() => {
                  const current = instances.find(i => i.name === selected)
                  return (
                    <div className="relative">
                      <select
                        value={selected}
                        onChange={e => setSelected(e.target.value)}
                        className="appearance-none h-8 pl-6 pr-7 rounded-md border bg-background text-xs font-medium focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer max-w-[120px] sm:max-w-none"
                      >
                        {instances.map(inst => (
                          <option key={inst.name} value={inst.name}>{inst.name}</option>
                        ))}
                      </select>
                      <span className={cn(
                        'pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 w-1.5 h-1.5 rounded-full',
                        current?.online ? 'bg-emerald-500' : 'bg-red-500'
                      )} />
                      <svg className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" /></svg>
                    </div>
                  )
                })()}
              </div>
            )}
            <button
              onClick={logout}
              className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs text-muted-foreground hover:text-foreground hover:bg-accent transition-colors"
            >
              <LogOut className="w-3.5 h-3.5" />
              <span className="hidden sm:inline">退出</span>
            </button>
          </div>
        </div>
      </header>

      <div className="flex flex-1 max-w-screen-xl mx-auto w-full">
        <aside className="hidden md:block w-48 shrink-0 border-r py-6 px-3">
          <nav className="space-y-1">
            {NAV.map(item => (
              <button
                key={item.id}
                onClick={() => setPage(item.id)}
                className={cn(
                  'w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors',
                  page === item.id
                    ? 'bg-accent font-medium text-foreground'
                    : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                )}
              >
                {item.icon}
                {item.label}
              </button>
            ))}
          </nav>
        </aside>

        <main className="flex-1 p-4 sm:p-6 min-w-0 pb-20 md:pb-6">
          {!selected && page !== 'settings' && page !== 'logs' && (
            <div className="flex items-center justify-center h-64 text-muted-foreground">
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
      <nav className="md:hidden fixed bottom-0 left-0 right-0 z-10 border-t bg-background/95 backdrop-blur-sm">
        <div className="flex items-stretch h-14">
          {NAV.map(item => (
            <button
              key={item.id}
              onClick={() => setPage(item.id)}
              className={cn(
                'flex-1 flex flex-col items-center justify-center gap-0.5 text-[10px] transition-colors',
                page === item.id
                  ? 'text-primary'
                  : 'text-muted-foreground'
              )}
            >
              <span className={cn('w-5 h-5', page === item.id && '[&>svg]:stroke-[2.5]')}>
                {item.icon}
              </span>
              {item.label}
            </button>
          ))}
        </div>
      </nav>
    </div>
  )
}
