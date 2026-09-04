'use client'

import { useMemo, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { useRouter } from 'next/navigation'
import { ChevronDown, ChevronLeft, ChevronRight, CircleHelp, FileJson, LayoutDashboard, ListFilter, Menu, RefreshCw, Search, TerminalSquare, X } from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'


const baseRows = [
  { id: 'unknown-2241', time: '22:39:42', raw: 'AUTH_FAIL :: usr#9287 :: src=10.20.30.4', source: 'auth-service', status: 'Analyzing', format: '—', confidence: '—' },
  { id: 'unknown-2240', time: '22:35:21', raw: 'XYZ code=E42 node=server-4', source: 'xyz-service', status: 'Format Detected', format: 'Custom Event', confidence: '94%' },
  { id: 'unknown-2239', time: '22:31:08', raw: '[CUSTOM] req#821 status=blocked', source: 'gateway', status: 'Format Detected', format: 'Gateway Security', confidence: '89%' },
  { id: 'unknown-2238', time: '22:27:51', raw: 'evt@unknown type=payment_failure user=481', source: 'payment-service', status: 'Analyzing', format: '—', confidence: '—' },
  { id: 'unknown-2237', time: '22:24:17', raw: 'node=server-7 :: action=restricted :: code=R41', source: 'security-service', status: 'Format Detected', format: 'Security Event', confidence: '91%' },
]
const rows = Array(12).fill(baseRows).flat().map((row, i) => ({ ...row, id: `${row.id}-${i}` }))



function SelectFilter({ label, value, options, onChange }) {
  return <label className="relative min-w-32"><span className="sr-only">{label}</span><select aria-label={label} value={value} onChange={(event) => onChange(event.target.value)} className="h-9 w-full appearance-none rounded-md border border-input bg-background px-3 pr-8 text-xs text-foreground outline-none focus:ring-1 focus:ring-ring">{options.map((option) => <option key={option}>{option}</option>)}</select><ChevronDown className="pointer-events-none absolute right-2.5 top-3 size-3.5 text-muted-foreground" /></label>
}

export function UnknownLogs() {
  const router = useRouter()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('All Status')
  const [range, setRange] = useState('Last 24 hours')
  const [mobileOpen, setMobileOpen] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const filtered = useMemo(() => {
    setCurrentPage(1)
    return rows.filter((row) => `${row.raw} ${row.source} ${row.status} ${row.format}`.toLowerCase().includes(search.toLowerCase()) && (status === 'All Status' || row.status === status))
  }, [search, status])

  const totalPages = Math.ceil(filtered.length / pageSize)
  const paginated = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const clear = () => { setSearch(''); setStatus('All Status'); setRange('Last 24 hours') }
  return <div className="min-h-screen bg-background text-foreground"><div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>{mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><button className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} aria-label="Close navigation" /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}<main className="md:pl-60"><div className="mx-auto max-w-[1500px] px-5 py-5 sm:px-8 sm:py-7"><header className="mb-7 flex items-start justify-between gap-4"><div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Workspace</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">Unknown Logs</h1><p className="mt-1 text-sm text-muted-foreground">Logs that could not be matched to an existing parser</p></div></div><div className="flex items-center gap-2"><ThemeToggle /><span className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 text-xs text-muted-foreground"><span className="size-1.5 rounded-full bg-emerald-400" />Live</span><Button variant="outline" size="sm" className="gap-2" onClick={() => { setRefreshing(true); setTimeout(() => setRefreshing(false), 700) }}><RefreshCw className={refreshing ? 'animate-spin' : ''} />Refresh</Button></div></header><section aria-label="Unknown log metrics" className="mb-5 grid grid-cols-3 divide-x divide-border rounded-lg border border-border bg-card/40"><div className="px-4 py-4"><p className="text-[11px] uppercase tracking-wider text-muted-foreground">Unknown Logs</p><p className="mt-1 font-mono text-xl">630</p></div><div className="px-4 py-4"><p className="text-[11px] uppercase tracking-wider text-muted-foreground">Analyzing</p><p className="mt-1 font-mono text-xl text-amber-400">18</p></div><div className="px-4 py-4"><p className="text-[11px] uppercase tracking-wider text-muted-foreground">Formats Detected</p><p className="mt-1 font-mono text-xl text-emerald-400">42</p></div></section><section aria-label="Unknown log filters" className="mb-5 flex flex-col gap-3 rounded-lg border border-border bg-card/40 p-3"><div className="relative"><Search className="absolute left-3 top-2.5 size-4 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search unknown logs..." className="h-9 border-input bg-background pl-9 text-sm" /></div><div className="flex flex-wrap items-center gap-2"><SelectFilter label="Status" value={status} options={['All Status', 'Analyzing', 'Format Detected', 'Failed']} onChange={setStatus} /><SelectFilter label="Time Range" value={range} options={['Last 15 minutes', 'Last hour', 'Last 24 hours', 'Last 7 days']} onChange={setRange} /><Button variant="ghost" size="sm" className="text-xs text-muted-foreground" onClick={clear}>Clear filters</Button></div></section><div className="overflow-hidden rounded-lg border border-border"><div className="overflow-x-auto"><Table><TableHeader><TableRow><TableHead>Time</TableHead><TableHead>Raw Log</TableHead><TableHead>Source</TableHead><TableHead>Status</TableHead><TableHead>Detected Format</TableHead><TableHead>Confidence</TableHead><TableHead className="text-right">Action</TableHead></TableRow></TableHeader><TableBody>{paginated.map((row) => <TableRow key={row.id} tabIndex={0} className="cursor-pointer" onClick={() => router.push(`/unknown-logs/${row.id}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') router.push(`/unknown-logs/${row.id}`) }}><TableCell className="font-mono text-xs text-muted-foreground">{row.time}</TableCell><TableCell className="min-w-64 font-mono text-xs text-foreground">{row.raw}</TableCell><TableCell className="font-mono text-xs text-muted-foreground">{row.source}</TableCell><TableCell><span className={`inline-flex items-center gap-2 text-xs ${row.status === 'Analyzing' ? 'text-amber-400' : 'text-emerald-400'}`}><span className={`size-1.5 rounded-full ${row.status === 'Analyzing' ? 'animate-pulse bg-amber-400' : 'bg-emerald-400'}`} />{row.status}</span></TableCell><TableCell className="text-xs text-muted-foreground">{row.format}</TableCell><TableCell className="font-mono text-xs text-muted-foreground">{row.confidence}</TableCell><TableCell className="text-right text-xs text-muted-foreground">{row.status === 'Format Detected' ? <button className="text-foreground underline-offset-4 hover:underline" onClick={(event) => { event.stopPropagation(); router.push(`/unknown-logs/${row.id}`) }}>Review →</button> : <span>Analyzing...</span>}</TableCell></TableRow>)}</TableBody></Table></div></div><footer className="flex flex-col gap-3 py-4 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between"><span>Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filtered.length)} of {filtered.length} unknown logs</span><div className="flex items-center gap-1"><Button variant="outline" size="icon" className="size-8" aria-label="Previous page" disabled={currentPage === 1} onClick={() => setCurrentPage(p => Math.max(1, p - 1))}><ChevronLeft className="size-4" /></Button><div className="flex items-center gap-1 mx-2 text-sm font-medium">Page {currentPage} of {totalPages || 1}</div><Button variant="outline" size="icon" className="size-8" aria-label="Next page" disabled={currentPage === totalPages || totalPages === 0} onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}><ChevronRight className="size-4" /></Button><SelectFilter label="Page size" value={`${pageSize} per page`} options={['10 per page', '25 per page', '50 per page']} onChange={(val) => { setPageSize(Number(val.split(' ')[0])); setCurrentPage(1); }} /></div></footer></div></main></div>
}
