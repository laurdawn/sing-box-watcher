import React from 'react'

// 标准 16 色 ANSI 调色板（适配暗色终端风格）
const ANSI_COLORS: Record<number, string> = {
  30: '#4e4e4e', 31: '#ff5f5f', 32: '#5faf5f', 33: '#d7af5f',
  34: '#5f87d7', 35: '#af5faf', 36: '#5fafaf', 37: '#d0d0d0',
  // 亮色
  90: '#808080', 91: '#ff8787', 92: '#87d787', 93: '#ffd787',
  94: '#87afd7', 95: '#d787d7', 96: '#87d7d7', 97: '#ffffff',
}

interface Span { text: string; color?: string; bg?: string; bold?: boolean }

export function parseAnsi(raw: string): Span[] {
  const spans: Span[] = []
  // eslint-disable-next-line no-control-regex
  const re = /\x1b\[([0-9;]*)m/g
  let last = 0
  let color: string | undefined
  let bg: string | undefined
  let bold = false

  const flush = (text: string) => {
    if (text) spans.push({ text, color, bg, bold })
  }

  let m: RegExpExecArray | null
  while ((m = re.exec(raw)) !== null) {
    flush(raw.slice(last, m.index))
    last = m.index + m[0].length

    const codes = m[1] === '' ? [0] : m[1].split(';').map(Number)
    let i = 0
    while (i < codes.length) {
      const c = codes[i]
      if (c === 0) { color = undefined; bg = undefined; bold = false }
      else if (c === 1) { bold = true }
      else if (c === 22) { bold = false }
      else if (c >= 30 && c <= 37) { color = ANSI_COLORS[c] }
      else if (c >= 90 && c <= 97) { color = ANSI_COLORS[c] }
      else if (c === 39) { color = undefined }
      else if (c >= 40 && c <= 47) { bg = ANSI_COLORS[c - 10] }
      else if (c === 49) { bg = undefined }
      else if (c === 38 && codes[i + 1] === 5) {
        // 256 色前景
        color = ansi256(codes[i + 2])
        i += 2
      } else if (c === 48 && codes[i + 1] === 5) {
        bg = ansi256(codes[i + 2])
        i += 2
      }
      i++
    }
  }
  flush(raw.slice(last))
  return spans
}

function ansi256(n: number): string {
  if (n < 16) return ANSI_COLORS[n < 8 ? n + 30 : n + 82] ?? '#888'
  if (n >= 232) { const v = 8 + (n - 232) * 10; return `rgb(${v},${v},${v})` }
  n -= 16
  const b = n % 6; n = Math.floor(n / 6)
  const g = n % 6; const r = Math.floor(n / 6)
  const c = (x: number) => x === 0 ? 0 : 55 + x * 40
  return `rgb(${c(r)},${c(g)},${c(b)})`
}

export function AnsiText({ raw, keyword }: { raw: string; keyword?: string }) {
  const spans = parseAnsi(raw)
  return (
    <>
      {spans.map((s, i) => {
        const style: React.CSSProperties = {}
        if (s.color) style.color = s.color
        if (s.bg) style.backgroundColor = s.bg
        if (s.bold) style.fontWeight = 'bold'

        if (keyword && s.text.toLowerCase().includes(keyword.toLowerCase())) {
          const idx = s.text.toLowerCase().indexOf(keyword.toLowerCase())
          return (
            <span key={i} style={style}>
              {s.text.slice(0, idx)}
              <mark className="bg-amber-200 dark:bg-amber-700 text-foreground rounded px-0.5">
                {s.text.slice(idx, idx + keyword.length)}
              </mark>
              {s.text.slice(idx + keyword.length)}
            </span>
          )
        }
        return <span key={i} style={style}>{s.text}</span>
      })}
    </>
  )
}
