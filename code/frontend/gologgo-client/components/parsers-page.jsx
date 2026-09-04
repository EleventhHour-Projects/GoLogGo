'use client'

import { useMemo, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { useRouter } from 'next/navigation'
import { ChevronDown, CircleHelp, FileJson, LayoutDashboard, ListFilter, Menu, Plus, Search, TerminalSquare, X } from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'


const parsers = [
  { name: 'AuthParser', format: 'Authentication', status: 'Active', logs: '5,241', lastUsed: 'Just now', created: 'Aug 20, 2026' },
  { name: 'NginxParser', format: 'Nginx', status: 'Active', logs: '8,421', lastUsed: '2 min ago', created: 'Aug 18, 2026' },
  { name: 'PaymentParser', format: 'Payment', status: 'Active', logs: '4,120', lastUsed: '5 min ago', created: 'Aug 22, 2026' },
  { name: 'PostgresParser', format: 'PostgreSQL', status: 'Active', logs: '3,241', lastUsed: '8 min ago', created: 'Aug 15, 2026' },
  { name: 'UserParser', format: 'User Activity', status: 'Active', logs: '1,698', lastUsed: '12 min ago', created: 'Aug 21, 2026' },
  { name: 'GatewayParser', format: 'Gateway Security', status: 'Active', logs: '780', lastUsed: '18 min ago', created: 'Aug 28, 2026' },
  { name: 'LegacyParser', format: 'Legacy Application', status: 'Inactive', logs: '391', lastUsed: '2 days ago', created: 'Aug 02, 2026' },
  { name: 'SecurityParser', format: 'Security Event', status: 'Active', logs: '612', lastUsed: '25 min ago', created: 'Aug 27, 2026' },
]



function SelectFilter({ label, value, onChange, options }) {
  return <label className="relative flex min-w-32 items-center"><span className="sr-only">{label}</span><select aria-label={label} value={value} onChange={(event) => onChange(event.target.value)} className="h-9 w-full appearance-none rounded-md border border-input bg-background px-3 pr-8 text-xs text-foreground outline-none focus:ring-1 focus:ring-ring">{options.map((option) => <option key={option}>{option}</option>)}</select><ChevronDown className="pointer-events-none absolute right-2.5 size-3.5 text-muted-foreground" /></label>
}

export function ParsersPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('All')
  const [format, setFormat] = useState('All Formats')
  const [mobileOpen, setMobileOpen] = useState(false)
  const router = useRouter()
  const filtered = useMemo(() => parsers.filter((parser) => `${parser.name} ${parser.format}`.toLowerCase().includes(search.toLowerCase()) && (status === 'All' || parser.status === status) && (format === 'All Formats' || parser.format === format)), [search, status, format])
  const clear = () => { setSearch(''); setStatus('All'); setFormat('All Formats') }

  return <div className="min-h-screen bg-background text-foreground"><div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>{mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><button className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} aria-label="Close navigation" /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}<main className="md:pl-60"><div className="mx-auto max-w-[1500px] px-5 py-5 sm:px-8 sm:py-7"><header className="mb-7 flex items-start justify-between gap-4"><div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Workspace</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">Parsers</h1><p className="mt-1 text-sm text-muted-foreground">Manage and monitor available log parsers</p></div></div><div className="flex items-center gap-2"><ThemeToggle /><Button size="sm" className="gap-2"><Plus className="size-3.5" />Create Parser</Button></div></header>
    <section aria-label="Parser metrics" className="mb-5 grid grid-cols-1 gap-3 sm:grid-cols-3"><Card className="py-4"><CardContent className="px-4"><p className="text-xs text-muted-foreground">TOTAL PARSERS</p><p className="mt-2 font-mono text-xl font-medium">12</p></CardContent></Card><Card className="py-4"><CardContent className="px-4"><p className="text-xs text-muted-foreground">ACTIVE PARSERS</p><p className="mt-2 font-mono text-xl font-medium">11</p></CardContent></Card><Card className="py-4"><CardContent className="px-4"><p className="text-xs text-muted-foreground">LOGS PROCESSED</p><p className="mt-2 font-mono text-xl font-medium">23,891</p></CardContent></Card></section>
    <section aria-label="Parser filters" className="mb-4 flex flex-col gap-3 rounded-lg border border-border bg-card/40 p-3 sm:flex-row sm:items-center"><div className="relative min-w-56 flex-1"><Search className="absolute left-3 top-2.5 size-4 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search parsers..." className="h-9 border-input bg-background pl-9 text-sm" /></div><div className="flex flex-wrap items-center gap-2"><SelectFilter label="Status" value={status} onChange={setStatus} options={['All', 'Active', 'Inactive']} /><SelectFilter label="Format" value={format} onChange={setFormat} options={['All Formats', ...Array.from(new Set(parsers.map((parser) => parser.format)))]} /><Button variant="ghost" size="sm" className="text-xs text-muted-foreground" onClick={clear}>Clear filters</Button></div></section>
    <Card><CardContent className="p-0"><div className="overflow-x-auto"><Table><TableHeader><TableRow><TableHead>PARSER</TableHead><TableHead>FORMAT</TableHead><TableHead>STATUS</TableHead><TableHead>LOGS PROCESSED</TableHead><TableHead>LAST USED</TableHead><TableHead>CREATED</TableHead><TableHead className="text-right">ACTION</TableHead></TableRow></TableHeader><TableBody>{filtered.map((parser) => <TableRow key={parser.name} tabIndex={0} className="cursor-pointer" onClick={() => router.push(`/parsers/${parser.name.toLowerCase()}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') router.push(`/parsers/${parser.name.toLowerCase()}`) }}><TableCell className="font-mono text-xs font-medium">{parser.name}</TableCell><TableCell className="text-sm text-muted-foreground">{parser.format}</TableCell><TableCell><Badge variant="outline" className={parser.status === 'Active' ? 'border-emerald-500/30 text-emerald-400' : 'border-border text-muted-foreground'}><span className={`mr-1.5 inline-block size-1.5 rounded-full ${parser.status === 'Active' ? 'bg-emerald-400' : 'bg-muted-foreground'}`} />{parser.status}</Badge></TableCell><TableCell className="font-mono text-xs">{parser.logs}</TableCell><TableCell className="text-sm text-muted-foreground">{parser.lastUsed}</TableCell><TableCell className="text-sm text-muted-foreground">{parser.created}</TableCell><TableCell className="text-right"><span className="text-xs text-muted-foreground">View <span aria-hidden="true">→</span></span></TableCell></TableRow>)}</TableBody></Table></div><div className="border-t border-border px-5 py-3 text-xs text-muted-foreground">Showing 1–{filtered.length} of 12 parsers</div></CardContent></Card>
  </div></main></div>
}
