import { useState, useEffect, useRef } from 'react'
import { Trash2, PauseCircle, PlayCircle, Search, X } from 'lucide-react'
import { useLogs } from '@/hooks/useLogs'
import { AnsiText } from '@/lib/ansi'
import { cn } from '@/lib/utils'

interface Props {
  instance: string
}

type Level = 'ALL' | 'TRACE' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL'

const LEVELS: Level[] = ['ALL', 'TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL']

const levelColor: Record<string, string> = {
  TRACE:   'text-muted-foreground',
  DEBUG:   'text-blue-400',
  INFO:    'text-foreground',
  WARN:    'text-amber-400',
  ERROR:   'text-red-400',
  FATAL:   'text-red-500 font-bold',
  PANIC:   'text-red-500 font-bold',
}

const levelPriority: Record<string, number> = {
  TRACE: 0, DEBUG: 1, INFO: 2, WARN: 3, ERROR: 4, FATAL: 5, PANIC: 6,
}

export function Logs({ instance }: Props) {
  const { logs, clear } = useLogs(instance)
  const [minLevel, setMinLevel] = useState<Level>('INFO')
  const [keyword, setKeyword] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const bottomRef = useRef<HTMLDivElement>(null)

  const filtered = logs.filter(l => {
    if (minLevel !== 'ALL' && (levelPriority[l.level] ?? 0) < (levelPriority[minLevel] ?? 0)) return false
    if (keyword && !l.message.toLowerCase().includes(keyword.toLowerCase())) return false
    return true
  })

  useEffect(() => {
    if (autoScroll) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [filtered.length, autoScroll])

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)]">
      {/* 工具栏 */}
      <div className="flex flex-col gap-2 mb-3">
        <div className="flex items-center gap-2 flex-wrap">
          {/* 级别筛选 */}
          <div className="flex items-center gap-1 rounded-lg border bg-card p-1 overflow-x-auto">
            {LEVELS.map(l => (
              <button
                key={l}
                onClick={() => setMinLevel(l)}
                className={cn(
                  'px-2 sm:px-2.5 py-1 text-xs rounded-md transition-colors shrink-0 font-mono font-medium',
                  minLevel === l
                    ? 'bg-blue-600 text-white'
                    : 'text-muted-foreground hover:bg-accent'
                )}
              >
                {l}
              </button>
            ))}
          </div>

          <button
            onClick={() => setAutoScroll(v => !v)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border bg-card hover:bg-accent transition-colors shrink-0"
          >
            {autoScroll
              ? <><PauseCircle className="w-3.5 h-3.5 text-blue-500" /> 暂停</>
              : <><PlayCircle className="w-3.5 h-3.5 text-emerald-500" /> 滚动</>
            }
          </button>

          <div className="relative flex-1 min-w-[120px]">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
            <input
              value={keyword}
              onChange={e => setKeyword(e.target.value)}
              placeholder="关键字过滤"
              className="h-8 pl-8 pr-7 rounded-lg border bg-card text-xs focus:outline-none focus:ring-1 focus:ring-blue-500 w-full"
            />
            {keyword && (
              <button onClick={() => setKeyword('')} className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground">
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          <button
            onClick={clear}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border bg-card hover:bg-accent text-muted-foreground transition-colors shrink-0"
          >
            <Trash2 className="w-3.5 h-3.5" /> 清空
          </button>

          <span className="text-xs text-muted-foreground shrink-0 font-mono">
            {filtered.length} / {logs.length}
          </span>
        </div>
      </div>

      {/* 日志列表 */}
      <div className="flex-1 overflow-y-auto rounded-xl border bg-slate-950 dark:bg-black font-mono text-xs">
        {filtered.length === 0 && (
          <div className="flex items-center justify-center h-32 text-slate-500">
            等待日志...
          </div>
        )}
        <div className="p-4 space-y-0.5">
          {filtered.map((log, i) => (
            <div key={i} className="flex gap-3 leading-5 hover:bg-white/5 px-1.5 py-0.5 rounded transition-colors">
              <span className={cn('shrink-0 w-11 text-[10px] uppercase tracking-wide font-semibold', levelColor[log.level] ?? 'text-slate-300')}>
                {log.level.slice(0, 5)}
              </span>
              <span className="text-slate-300 break-all">
                <AnsiText raw={log.message} keyword={keyword} />
              </span>
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </div>
    </div>
  )
}

