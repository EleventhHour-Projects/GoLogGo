'use client'

import { useMemo, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { useRouter } from 'next/navigation'
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleHelp,
  FileJson,
  LayoutDashboard,
  ListFilter,
  Menu,
  RefreshCw,
  Search,
  TerminalSquare,
  X,
} from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'


const baseLogs = [
  ['22:41:03', 'ERROR', 'auth-service', 'User authentication failed', 'Authentication', 'AuthParser', 'Parsed'],
  ['22:40:15', 'INFO', 'payment-service', 'Payment completed successfully', 'Payment', 'PaymentParser', 'Parsed'],
  ['22:39:42', 'WARN', 'nginx', 'Invalid request received', 'Unknown', '—', 'Unknown'],
  ['22:38:51', 'INFO', 'user-service', 'New user registered', 'User Activity', 'UserParser', 'Parsed'],
  ['22:37:22', 'ERROR', 'database', 'Connection timeout', 'PostgreSQL', 'PostgresParser', 'Parsed'],
  ['22:36:14', 'INFO', 'auth-service', 'User session created', 'Authentication', 'AuthParser', 'Parsed'],
  ['22:35:08', 'WARN', 'nginx', 'Request rate limit approaching', 'Nginx', 'NginxParser', 'Parsed'],
  ['22:34:51', 'ERROR', 'payment-service', 'Payment gateway unavailable', 'Payment', 'PaymentParser', 'Parsed'],
  ['22:33:27', 'INFO', 'user-service', 'Profile updated', 'User Activity', 'UserParser', 'Parsed'],
  ['22:32:10', 'WARN', 'database', 'Slow query detected', 'PostgreSQL', 'PostgresParser', 'Parsed'],
]
const logs = Array(12).fill(baseLogs).flat().map((log, i) => [...log, i])

const options = {
  Status: ['All', 'Parsed', 'Unknown'],
  Level: ['All', 'INFO', 'WARN', 'ERROR'],
  Service: ['All', 'auth-service', 'payment-service', 'nginx', 'user-service', 'database'],
  Parser: ['All', 'AuthParser', 'NginxParser', 'PaymentParser', 'PostgresParser', 'Unknown'],
  'Time range': ['Last 15 minutes', 'Last hour', 'Last 24 hours', 'Last 7 days'],
}



function FilterSelect({ label, value, onChange }) {
  return <label className="relative flex min-w-[128px] flex-1 items-center"><span className="sr-only">{label}</span><select aria-label={label} value={value} onChange={(event) => onChange(event.target.value)} className="h-9 w-full appearance-none rounded-md border border-input bg-background px-3 pr-8 text-xs text-foreground outline-none transition-colors focus:ring-1 focus:ring-ring">{options[label].map((option) => <option key={option}>{option}</option>)}</select><ChevronDown className="pointer-events-none absolute right-2.5 size-3.5 text-muted-foreground" /></label>
}

export function LogsExplorer() {
  const router = useRouter()
  const [search, setSearch] = useState('')
  const [filters, setFilters] = useState({ Status: 'All', Level: 'All', Service: 'All', Parser: 'All', 'Time range': 'Last 24 hours' })
  const [mobileOpen, setMobileOpen] = useState(false)
  const [refreshed, setRefreshed] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const filteredLogs = useMemo(() => {
    setCurrentPage(1)
    return logs.filter((log) => log.join(' ').toLowerCase().includes(search.toLowerCase()) && (filters.Status === 'All' || log[6] === filters.Status) && (filters.Level === 'All' || log[1] === filters.Level) && (filters.Service === 'All' || log[2] === filters.Service) && (filters.Parser === 'All' || log[5] === filters.Parser))
  }, [search, filters])

  const totalPages = Math.ceil(filteredLogs.length / pageSize)
  const paginatedLogs = filteredLogs.slice((currentPage - 1) * pageSize, currentPage * pageSize)
  const clearFilters = () => { setSearch(''); setFilters({ Status: 'All', Level: 'All', Service: 'All', Parser: 'All', 'Time range': 'Last 24 hours' }) }

  return <div className="min-h-screen bg-background text-foreground"><div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>{mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><button className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} aria-label="Close navigation" /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}<main className="md:pl-60"><div className="mx-auto max-w-[1500px] px-5 py-5 sm:px-8 sm:py-7"><header className="mb-7 flex items-start justify-between gap-4"><div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Workspace</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">Logs</h1><p className="mt-1 text-sm text-muted-foreground">Browse and investigate incoming log events</p></div></div><div className="flex items-center gap-2"><ThemeToggle /><span className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 text-xs text-muted-foreground"><span className="size-1.5 rounded-full bg-emerald-400" />Live</span><Button variant="outline" size="icon" className="size-8" onClick={() => { setRefreshed(true); setTimeout(() => setRefreshed(false), 700) }} aria-label="Refresh logs"><RefreshCw className={`size-3.5 ${refreshed ? 'animate-spin' : ''}`} /></Button></div></header><section aria-label="Log filters" className="mb-5 flex flex-col gap-3 rounded-lg border border-border bg-card/40 p-3"><div className="relative"><Search className="absolute left-3 top-2.5 size-4 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search logs..." className="h-9 border-input bg-background pl-9 text-sm" /></div><div className="flex flex-wrap items-center gap-2">{(Object.keys(options)).map((label) => <FilterSelect key={label} label={label} value={filters[label]} onChange={(value) => setFilters((current) => ({ ...current, [label]: value }))} />)}<button onClick={clearFilters} className="px-2 text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline">Clear filters</button></div></section><section className="overflow-hidden rounded-lg border border-border bg-card/20"><div className="flex items-center justify-between border-b border-border px-4 py-3"><div><h2 className="text-sm font-medium">All logs</h2><p className="mt-0.5 text-xs text-muted-foreground">24,521 events across 5 services</p></div><span className="font-mono text-xs text-muted-foreground">Auto-refresh on</span></div><div className="overflow-x-auto"><Table><TableHeader><TableRow className="hover:bg-transparent"><TableHead className="w-24 pl-4">Time</TableHead><TableHead>Level</TableHead><TableHead>Service</TableHead><TableHead className="min-w-64">Message</TableHead><TableHead>Format</TableHead><TableHead>Parser</TableHead><TableHead>Status</TableHead><TableHead className="w-8 pr-4" /></TableRow></TableHeader><TableBody>{paginatedLogs.map((log, index) => <TableRow key={`${log[0]}-${log[7] || index}`} tabIndex={0} className="cursor-pointer transition-colors hover:bg-accent/50" onClick={() => router.push(`/logs/log-${2241 - index}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') router.push(`/logs/log-${2241 - index}`) }}><TableCell className="pl-4 font-mono text-xs text-muted-foreground">{log[0]}</TableCell><TableCell><Badge variant="outline" className={`font-mono text-[10px] ${log[1] === 'ERROR' ? 'border-destructive/40 bg-destructive/10 text-destructive' : log[1] === 'WARN' ? 'border-amber-500/30 bg-amber-500/10 text-amber-400' : 'text-muted-foreground'}`}>{log[1]}</Badge></TableCell><TableCell className="font-mono text-xs">{log[2]}</TableCell><TableCell className="text-sm">{log[3]}</TableCell><TableCell className="text-xs text-muted-foreground">{log[4]}</TableCell><TableCell className="font-mono text-xs text-muted-foreground">{log[5]}</TableCell><TableCell><span className={`inline-flex items-center gap-1.5 text-xs ${log[6] === 'Unknown' ? 'text-amber-400' : 'text-emerald-400'}`}><span className={`size-1.5 rounded-full ${log[6] === 'Unknown' ? 'bg-amber-400' : 'bg-emerald-400'}`} />{log[6]}</span></TableCell><TableCell className="pr-4 text-right text-muted-foreground">›</TableCell></TableRow>)}</TableBody></Table></div></section><div className="flex flex-col gap-4 py-4 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between"><span>Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filteredLogs.length)} of {filteredLogs.length} logs</span><div className="flex items-center gap-1"><Button variant="outline" size="sm" className="h-8 gap-1" disabled={currentPage === 1} onClick={() => setCurrentPage(p => Math.max(1, p - 1))}><ChevronLeft className="size-3.5" />Previous</Button><div className="flex items-center gap-1 mx-2 text-sm font-medium">Page {currentPage} of {totalPages || 1}</div><Button variant="outline" size="sm" className="h-8 gap-1" disabled={currentPage === totalPages || totalPages === 0} onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}>Next<ChevronRight className="size-3.5" /></Button><select aria-label="Logs per page" value={pageSize} onChange={(e) => { setPageSize(Number(e.target.value)); setCurrentPage(1); }} className="ml-2 h-8 rounded-md border border-input bg-background px-2 text-xs"><option value={10}>10 per page</option><option value={25}>25 per page</option><option value={50}>50 per page</option></select></div></div></div></main></div>
}
