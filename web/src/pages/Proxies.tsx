import { RefreshCw } from 'lucide-react'
import { useGroups } from '@/hooks/useGroups'
import { Group, GroupItem } from '@/lib/api'
import { cn } from '@/lib/utils'

interface Props {
  instance: string
}

export function Proxies({ instance }: Props) {
  const { snapshot, selectOutbound, urlTest, testing } = useGroups(instance)

  if (!snapshot || !snapshot.groups || snapshot.groups.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
        暂无代理分组数据，请确认 sing-box 已启动...
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {snapshot.groups.map(group => (
        <GroupCard
          key={group.tag}
          group={group}
          onSelect={(outboundTag) => selectOutbound(group.tag, outboundTag)}
          onURLTest={() => urlTest(group.tag)}
          testing={testing.has(group.tag)}
        />
      ))}
    </div>
  )
}

function GroupCard({
  group,
  onSelect,
  onURLTest,
  testing,
}: {
  group: Group
  onSelect: (tag: string) => void
  onURLTest: () => void
  testing: boolean
}) {
  const isSelector = group.type === 'selector'
  const isURLTest = group.type === 'urltest'

  return (
    <div className="rounded-xl border bg-card shadow-sm overflow-hidden">
      <div className="flex items-center justify-between px-5 py-3 border-b">
        <div className="flex items-center gap-2">
          <span className="font-medium">{group.tag}</span>
          <span className="text-xs text-muted-foreground px-2 py-0.5 rounded-md bg-muted">
            {group.type}
          </span>
          {group.selected && (
            <span className="text-xs text-emerald-600 dark:text-emerald-400">
              → {group.selected}
            </span>
          )}
        </div>
        {(isSelector || isURLTest) && (
          <button
            onClick={onURLTest}
            disabled={testing}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-md border hover:bg-accent transition-colors disabled:opacity-50"
          >
            <RefreshCw className={cn('w-3 h-3', testing && 'animate-spin')} />
            测速
          </button>
        )}
      </div>
      <div className="p-3 grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-2">
        {(group.items || []).map(item => (
          <NodeItem
            key={item.tag}
            item={item}
            selected={group.selected === item.tag}
            selectable={isSelector}
            onClick={() => isSelector && onSelect(item.tag)}
          />
        ))}
      </div>
    </div>
  )
}

function NodeItem({
  item,
  selected,
  selectable,
  onClick,
}: {
  item: GroupItem
  selected: boolean
  selectable: boolean
  onClick: () => void
}) {
  const delay = item.urlTestDelay
  const delayColor =
    delay === 0 ? 'text-muted-foreground' :
    delay < 200 ? 'text-emerald-500' :
    delay < 500 ? 'text-amber-500' :
    'text-red-500'

  return (
    <div
      onClick={onClick}
      className={cn(
        'rounded-lg border px-3 py-2.5 text-sm transition-colors',
        selectable ? 'cursor-pointer hover:bg-accent' : 'cursor-default',
        selected && 'border-primary bg-primary/5 dark:bg-primary/10',
      )}
    >
      <div className="flex items-center justify-between gap-1 mb-0.5">
        <span className={cn('text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground font-mono', selected && 'bg-primary/10')}>
          {item.type || 'unknown'}
        </span>
        {selected && (
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0" />
        )}
      </div>
      <div className="truncate text-xs font-medium mt-1" title={item.tag}>
        {item.tag}
      </div>
      <div className={cn('text-xs mt-0.5', delayColor)}>
        {delay > 0 ? `${delay} ms` : '未测试'}
      </div>
    </div>
  )
}
