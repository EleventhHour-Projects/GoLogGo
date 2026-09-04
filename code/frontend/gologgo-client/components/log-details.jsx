'use client'

import { useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { useRouter } from 'next/navigation'
import { ArrowLeft, Check, CircleHelp, Copy, FileJson, LayoutDashboard, ListFilter, Menu, TerminalSquare, X } from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'




const summary = [['Timestamp', '31 Aug 2026, 22:41:03'], ['Level', 'ERROR'], ['Service', 'auth-service'], ['Format', 'Authentication'], ['Parser', 'AuthParser'], ['Status', 'Parsed'], ['Log ID', 'log_8f32a91']]
const rawLog = '2026-08-31 22:41:03 ERROR auth-service:\nUser 9287 authentication failed from 10.20.30.4'
const normalized = '{\n  "timestamp": "2026-08-31T22:41:03",\n  "level": "ERROR",\n  "service": "auth-service",\n  "event": "authentication_failed",\n  "user": "9287",\n  "source_ip": "10.20.30.4"\n}'

const HighlightedRaw = () => (
  <>
    <span className="text-sky-400">2026-08-31 22:41:03</span> <span className="text-red-400 font-semibold">ERROR</span> <span className="text-amber-300">auth-service:</span><br />
    <span className="text-[#cccccc]">User </span><span className="text-[#b5cea8]">9287</span><span className="text-[#cccccc]"> authentication failed from </span><span className="text-[#ce9178]">10.20.30.4</span>
  </>
)

const HighlightedJSON = () => (
  <>
    <span className="text-[#ffd700]">{'{'}</span><br />
    {'  '}<span className="text-[#9cdcfe]">&quot;timestamp&quot;</span><span className="text-[#cccccc]">: </span><span className="text-[#ce9178]">&quot;2026-08-31T22:41:03&quot;</span><span className="text-[#cccccc]">,</span><br />
    {'  '}<span className="text-[#9cdcfe]">&quot;level&quot;</span><span className="text-[#cccccc]">: </span><span className="text-[#ce9178]">&quot;ERROR&quot;</span><span className="text-[#cccccc]">,</span><br />
    {'  '}<span className="text-[#9cdcfe]">&quot;service&quot;</span><span className="text-[#cccccc]">: </span><span className="text-[#ce9178]">&quot;auth-service&quot;</span><span className="text-[#cccccc]">,</span><br />
    {'  '}<span className="text-[#9cdcfe]">&quot;event&quot;</span><span className="text-[#cccccc]">: </span><span className="text-[#ce9178]">&quot;authentication_failed&quot;</span><span className="text-[#cccccc]">,</span><br />
    {'  '}<span className="text-[#9cdcfe]">&quot;user&quot;</span><span className="text-[#cccccc]">: </span><span className="text-[#ce9178]">&quot;9287&quot;</span><span className="text-[#cccccc]">,</span><br />
    {'  '}<span className="text-[#9cdcfe]">&quot;source_ip&quot;</span><span className="text-[#cccccc]">: </span><span className="text-[#ce9178]">&quot;10.20.30.4&quot;</span><br />
    <span className="text-[#ffd700]">{'}'}</span>
  </>
)

function CodePanel({ title, code, isRaw = false, isNormalized = false, onCopy }) {
  return <Card className="overflow-hidden"><CardHeader className="flex-row items-center justify-between border-b border-border px-4 py-3"><CardTitle className="text-sm font-medium">{title}</CardTitle><Button variant="ghost" size="sm" className="h-7 gap-2 text-xs text-muted-foreground" onClick={onCopy}><Copy className="size-3.5" />Copy</Button></CardHeader><CardContent className="overflow-x-auto bg-[#1e1e1e] p-0"><pre className="min-h-28 px-5 py-4 font-mono text-xs leading-6 sm:text-[13px]"><code>{isNormalized ? <HighlightedJSON /> : isRaw ? <HighlightedRaw /> : code}</code></pre></CardContent></Card>
}

export function LogDetails() {
  const router = useRouter()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [copied, setCopied] = useState('')
  const copy = async (name, value) => { await navigator.clipboard?.writeText(value); setCopied(name); setTimeout(() => setCopied(''), 1200) }
  return <div className="min-h-screen bg-background text-foreground"><div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>{mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><button className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} aria-label="Close navigation" /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}<main className="md:pl-60"><div className="mx-auto max-w-[1100px] px-5 py-5 sm:px-8 sm:py-7"><header className="mb-6 flex items-start justify-between gap-4"><div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><button onClick={() => router.push('/logs')} className="mb-3 flex items-center gap-2 text-xs text-muted-foreground transition-colors hover:text-foreground"><ArrowLeft className="size-3.5" />Back to Logs</button><h1 className="text-2xl font-semibold tracking-tight">Log Details</h1></div></div><div className="mt-8 flex items-center gap-2"><ThemeToggle /><Badge variant="outline" className="gap-2 border-emerald-500/30 text-emerald-400"><span className="size-1.5 rounded-full bg-emerald-400" />Parsed</Badge></div></header><section aria-labelledby="log-information" className="mb-5 border-y border-border py-4"><h2 id="log-information" className="mb-4 text-xs font-medium uppercase tracking-wider text-muted-foreground">Log Information</h2><div className="grid grid-cols-2 gap-x-6 gap-y-4 sm:grid-cols-4 lg:grid-cols-7">{summary.map(([label, value]) => <div key={label}><p className="text-[11px] text-muted-foreground">{label}</p><p className={`mt-1 truncate text-xs ${label === 'Level' ? 'font-mono text-red-400' : 'font-medium'}`}>{value}</p></div>)}</div></section><div className="flex flex-col gap-4"><CodePanel title="Raw Log" code={rawLog} isRaw onCopy={() => copy('raw', rawLog)} /><CodePanel title="Normalized Log" code={normalized} isNormalized onCopy={() => copy('normalized', normalized)} /></div><div className="mt-4 grid gap-4 lg:grid-cols-2"><Card><CardHeader className="border-b border-border px-4 py-3"><CardTitle className="text-sm font-medium">Detected Fields</CardTitle></CardHeader><CardContent className="grid gap-3 px-4 py-4">{[['event', 'authentication_failed'], ['user', '9287'], ['source_ip', '10.20.30.4'], ['level', 'ERROR'], ['service', 'auth-service']].map(([key, value]) => <div key={key} className="grid grid-cols-[100px_1fr] gap-3 text-xs"><span className="font-mono text-muted-foreground">{key}</span><span className="font-mono text-foreground">{value}</span></div>)}</CardContent></Card><Card><CardHeader className="border-b border-border px-4 py-3"><CardTitle className="text-sm font-medium">Parser Information</CardTitle></CardHeader><CardContent className="grid gap-3 px-4 py-4">{[['Parser', 'AuthParser'], ['Format', 'Authentication Log'], ['Status', 'Active'], ['Logs Processed', '5,241'], ['Last Used', 'Just now']].map(([label, value]) => <div key={label} className="flex items-center justify-between gap-4 text-xs"><span className="text-muted-foreground">{label}</span><span className="font-mono text-foreground">{value}</span></div>)}<Button variant="outline" size="sm" className="mt-2 w-fit text-xs" onClick={() => router.push('/parsers')}>View Parser <span aria-hidden="true">→</span></Button></CardContent></Card></div>{copied && <div role="status" className="fixed bottom-5 right-5 flex items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-xs shadow-lg"><Check className="size-3.5 text-emerald-400" />Copied {copied} log</div>}</div></main></div>
}
