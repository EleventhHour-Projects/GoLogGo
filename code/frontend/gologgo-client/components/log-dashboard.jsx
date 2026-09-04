'use client'

import { useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import {
  BarChart3,
  ChevronDown,
  CircleHelp,
  Database,
  FileJson,
  LayoutDashboard,
  ListFilter,
  Menu,
  Search,
  TerminalSquare,
  X,
} from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'


const parserData = [
  { name: 'Nginx', count: 8421 },
  { name: 'Apache', count: 5921 },
  { name: 'Auth', count: 5241 },
  { name: 'PostgreSQL', count: 3241 },
  { name: 'Docker', count: 1698 },
  { name: 'Custom', count: 780 },
]

const logs = [
  { id: 'log-2241', time: '22:41:03', level: 'ERROR', service: 'auth-service', message: 'User authentication failed', status: 'Parsed' },
  { id: 'log-2240', time: '22:40:15', level: 'INFO', service: 'payment-service', message: 'Payment completed', status: 'Parsed' },
  { id: 'log-2239', time: '22:39:42', level: 'WARN', service: 'nginx', message: 'Invalid request', status: 'Unknown' },
  { id: 'log-2238', time: '22:38:51', level: 'INFO', service: 'user-service', message: 'User registered', status: 'Parsed' },
  { id: 'log-2237', time: '22:37:22', level: 'ERROR', service: 'database', message: 'Connection timeout', status: 'Parsed' },
]



export function LogDashboard() {
  const router = useRouter()
  const [range, setRange] = useState('Last 24 hours')
  const [menuOpen, setMenuOpen] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const max = parserData[0].count

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>
      {mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><div className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}
      <main className="md:pl-60">
        <div className="mx-auto max-w-[1400px] px-5 py-5 sm:px-8 sm:py-7">
          <header className="mb-7 flex items-start justify-between gap-4">
            <div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Overview</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">Dashboard</h1><p className="mt-1 text-sm text-muted-foreground">Overview of your log processing system</p></div></div>
            <div className="flex items-center gap-2"><ThemeToggle /><Button variant="outline" size="sm" className="hidden gap-2 sm:flex" onClick={() => setMenuOpen(!menuOpen)}>{range}<ChevronDown className="size-3.5" />{menuOpen && <span className="absolute right-8 top-16 z-10 flex w-36 flex-col rounded-md border border-border bg-popover p-1 text-left shadow-lg"><button className="rounded px-2 py-1.5 text-left text-xs hover:bg-accent" onClick={() => { setRange('Last 24 hours'); setMenuOpen(false) }}>Last 24 hours</button><button className="rounded px-2 py-1.5 text-left text-xs hover:bg-accent" onClick={() => { setRange('Last 7 days'); setMenuOpen(false) }}>Last 7 days</button></span>}</Button></div>
          </header>

          <section aria-label="Log metrics" className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            {[['Total Logs', '24,521', '+12.4%', Database], ['Parsed Logs', '23,891', '+8.2%', BarChart3], ['Unknown Logs', '630', '-3.1%', CircleHelp], ['Error Logs', '42', '+2.4%', CircleHelp]].map(([label, value, trend, Icon]) => <Card key={label} className="gap-3 py-4"><CardContent className="px-4"><div className="flex items-center justify-between"><p className="text-xs text-muted-foreground">{label}</p><Icon className="size-4 text-muted-foreground/70" /></div><div className="mt-2 flex items-end justify-between gap-2"><p className="font-mono text-xl font-medium tracking-tight">{value}</p><span className="text-xs text-muted-foreground">{trend}</span></div></CardContent></Card>)}
          </section>

          <Card className="mt-5"><CardHeader className="flex-row items-center justify-between border-b border-border px-5 py-4"><div><CardTitle className="text-sm font-medium">Log Parsing Overview</CardTitle><p className="mt-1 text-xs text-muted-foreground">Successfully processed logs by parser</p></div><div className="relative"><Button variant="outline" size="sm" className="gap-2 sm:hidden" onClick={() => setMenuOpen(!menuOpen)}>{range}<ChevronDown className="size-3.5" /></Button><Button variant="outline" size="sm" className="hidden gap-2 sm:flex" onClick={() => setMenuOpen(!menuOpen)}>{range}<ChevronDown className="size-3.5" /></Button></div></CardHeader><CardContent className="px-5 py-6"><div className="flex flex-col gap-4">{parserData.map((item) => <div key={item.name} className="grid grid-cols-[80px_1fr_52px] items-center gap-3 text-sm sm:grid-cols-[110px_1fr_64px]"><span className="text-muted-foreground">{item.name}</span><div className="h-2 overflow-hidden rounded-sm bg-accent"><div className="h-full rounded-sm bg-primary/75" style={{ width: `${(item.count / max) * 100}%` }} /></div><span className="text-right font-mono text-xs text-muted-foreground">{item.count.toLocaleString()}</span></div>)}</div></CardContent></Card>

          <Card className="mt-5"><CardHeader className="flex-row items-center justify-between px-5 py-4"><div><CardTitle className="text-sm font-medium">Recent Logs</CardTitle><p className="mt-1 text-xs text-muted-foreground">Latest events received by Log.AI</p></div><Button variant="ghost" size="sm" className="text-xs text-muted-foreground hover:text-foreground" onClick={() => router.push('/logs')}>View all logs <span aria-hidden="true">→</span></Button></CardHeader><CardContent className="p-0"><div className="overflow-x-auto"><Table><TableHeader><TableRow><TableHead>Time</TableHead><TableHead>Level</TableHead><TableHead>Service</TableHead><TableHead>Message</TableHead><TableHead>Status</TableHead></TableRow></TableHeader><TableBody>{logs.map((log) => <TableRow key={log.id} tabIndex={0} className="cursor-pointer" onClick={() => router.push(`/logs/${log.id}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') router.push(`/logs/${log.id}`) }}><TableCell className="font-mono text-xs text-muted-foreground">{log.time}</TableCell><TableCell><Badge variant={log.level === 'ERROR' ? 'destructive' : 'secondary'} className={log.level === 'WARN' ? 'border-amber-500/30 bg-amber-500/10 text-amber-400' : 'font-mono text-[10px]'}>{log.level}</Badge></TableCell><TableCell className="font-mono text-xs">{log.service}</TableCell><TableCell className="min-w-48 text-sm">{log.message}</TableCell><TableCell><Badge variant="outline" className={log.status === 'Unknown' ? 'border-amber-500/30 text-amber-400' : 'border-emerald-500/30 text-emerald-400'}>{log.status}</Badge></TableCell></TableRow>)}</TableBody></Table></div></CardContent></Card>
        </div>
      </main>
    </div>
  )
}
