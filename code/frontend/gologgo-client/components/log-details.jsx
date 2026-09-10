'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { ArrowLeft, Check, Copy, Menu } from 'lucide-react'
import { Sidebar } from '@/components/sidebar'
import { ThemeToggle } from '@/components/theme-toggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

function HighlightedJSON({ code }) {
  const tokenPattern = /("(?:\\.|[^"\\])*"(?=\s*:)|"(?:\\.|[^"\\])*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|(true|false|null)|([{}[\],:])/g
  const parts = []
  let lastIndex = 0
  let match

  while ((match = tokenPattern.exec(code)) !== null) {
    if (match.index > lastIndex) parts.push(<span key={`text-${lastIndex}`}>{code.slice(lastIndex, match.index)}</span>)

    if (match[1]) {
      const isKey = /"\s*$/.test(match[1]) && code.slice(tokenPattern.lastIndex).match(/^\s*:/)
      parts.push(<span key={`string-${match.index}`} className={isKey ? 'text-[#9cdcfe]' : 'text-[#ce9178]'}>{match[1]}</span>)
    } else if (match[2]) {
      parts.push(<span key={`number-${match.index}`} className="text-[#b5cea8]">{match[2]}</span>)
    } else if (match[3]) {
      parts.push(<span key={`literal-${match.index}`} className="text-[#569cd6]">{match[3]}</span>)
    } else {
      parts.push(<span key={`punctuation-${match.index}`} className="text-[#ffd700]">{match[4]}</span>)
    }

    lastIndex = tokenPattern.lastIndex
  }

  if (lastIndex < code.length) parts.push(<span key={`text-${lastIndex}`}>{code.slice(lastIndex)}</span>)
  return <>{parts}</>
}

function CodePanel({ title, code, onCopy, isJSON = false }) {
  return (
    <Card className="overflow-hidden">
      <CardHeader className="flex-row items-center justify-between border-b border-border px-4 py-3">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Button variant="ghost" size="sm" className="h-7 gap-2 text-xs text-muted-foreground" onClick={onCopy}>
          <Copy className="size-3.5" />Copy
        </Button>
      </CardHeader>
      <CardContent className="overflow-x-auto bg-[#1e1e1e] p-0">
        <pre className="min-h-28 whitespace-pre-wrap px-5 py-4 font-mono text-xs leading-6 text-[#d4d4d4] sm:text-[13px]"><code>{isJSON ? <HighlightedJSON code={code} /> : code}</code></pre>
      </CardContent>
    </Card>
  )
}

export function LogDetails() {
  const router = useRouter()
  const { id } = useParams()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [copied, setCopied] = useState('')
  const [data, setData] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    fetch(`/api/logs/${id}`, { cache: 'no-store' })
      .then(async (response) => {
        const body = await response.json().catch(() => ({}))
        if (!response.ok) throw new Error(body.error || 'Failed to load log details')
        return body
      })
      .then(setData)
      .catch((loadError) => setError(loadError.message))
  }, [id])

  const request = data?.request
  const log = data?.logs?.[0]
  const normalized = log?.normalizedLog ? JSON.stringify(log.normalizedLog, null, 2) : ''
  const copy = async (name, value) => {
    await navigator.clipboard?.writeText(value)
    setCopied(name)
    setTimeout(() => setCopied(''), 1200)
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>
      {mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><button className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} aria-label="Close navigation" /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}
      <main className="md:pl-60">
        <div className="mx-auto max-w-[1100px] px-5 py-5 sm:px-8 sm:py-7">
          <header className="mb-6 flex items-start justify-between gap-4">
            <div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><button onClick={() => router.push('/logs')} className="mb-3 flex items-center gap-2 text-xs text-muted-foreground hover:text-foreground"><ArrowLeft className="size-3.5" />Back to Logs</button><h1 className="text-2xl font-semibold tracking-tight">Log Details</h1></div></div>
            <div className="mt-8 flex items-center gap-2"><ThemeToggle />{request && <Badge variant="outline" className="border-emerald-500/30 text-emerald-400">{request.status}</Badge>}</div>
          </header>
          {error && <p className="py-10 text-sm text-destructive">{error}</p>}
          {!error && !data && <p className="py-10 text-sm text-muted-foreground">Loading log details...</p>}
          {request && <>
            <section className="mb-5 border-y border-border py-4"><h2 className="mb-4 text-xs font-medium uppercase tracking-wider text-muted-foreground">Log Information</h2><div className="grid grid-cols-2 gap-4 sm:grid-cols-4"><div><p className="text-[11px] text-muted-foreground">Submitted</p><p className="mt-1 text-xs">{new Date(request.createdAt).toLocaleString()}</p></div><div><p className="text-[11px] text-muted-foreground">Status</p><p className="mt-1 text-xs">{request.status}</p></div><div><p className="text-[11px] text-muted-foreground">Attempts</p><p className="mt-1 font-mono text-xs">{request.attempts}</p></div><div><p className="text-[11px] text-muted-foreground">Log ID</p><p className="mt-1 truncate font-mono text-xs">{request.id}</p></div></div></section>
            <div className="flex flex-col gap-4"><CodePanel title="Raw Log" code={request.payload} onCopy={() => copy('raw', request.payload)} />{log && <CodePanel title="Normalized Log" code={normalized} isJSON onCopy={() => copy('normalized', normalized)} />}</div>
            {log?.hash && <p className="mt-4 font-mono text-xs text-muted-foreground">Fingerprint: {log.hash}</p>}
          </>}
          {copied && <div role="status" className="fixed bottom-5 right-5 flex items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-xs shadow-lg"><Check className="size-3.5 text-emerald-400" />Copied {copied} log</div>}
        </div>
      </main>
    </div>
  )
}
