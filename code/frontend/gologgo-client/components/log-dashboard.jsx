'use client'

import { useEffect, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { useRouter } from 'next/navigation'
import {
  BarChart3,
  CircleHelp,
  Database,
  Menu,
} from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'


export function LogDashboard() {
  const router = useRouter()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [rawLogs, setRawLogs] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)
  const [stats, setStats] = useState(null)
  const [parserData, setParserData] = useState([])
  const [recentLogs, setRecentLogs] = useState([])

  useEffect(() => {
    const loadDashboard = async () => {
      try {
        const [logsResponse, parsersResponse] = await Promise.all([
          fetch('/api/logs?limit=5&status=all', { cache: 'no-store' }),
          fetch('/api/parsers', { cache: 'no-store' }),
        ])

        if (logsResponse.ok) {
          const logsData = await logsResponse.json()
          setStats(logsData.stats)
          setRecentLogs((logsData.requests ?? []).map((request) => {
            const normalized = request.logs?.[0]?.normalizedLog ?? {}
            return {
              id: request.id,
              time: request.createdAt,
              level: normalized.severity || '—',
              service: normalized.service || '—',
              message: normalized.message || request.payload || '—',
              status: request.status,
            }
          }))
        }

        if (parsersResponse.ok) {
          const parsersData = await parsersResponse.json()
          const parsers = parsersData?.parsers ?? parsersData ?? []
          setParserData(parsers
            .map((parser) => ({ name: parser.name || parser.format || 'Unnamed', count: Number(parser.logsProcessed || 0) }))
            .filter((parser) => parser.count > 0)
            .sort((a, b) => b.count - a.count)
            .slice(0, 6))
        }
      } catch (error) {
        console.error('Failed to load dashboard data:', error)
      }
    }

    loadDashboard()
  }, [])

  const handleClearLogs = () => setRawLogs('')
  const handleProcessLogs = async () => {
    const logLines = rawLogs
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter(Boolean)

    if (logLines.length === 0) return
    setIsProcessing(true);
    try {
      const responses = await Promise.all(logLines.map((line) => fetch('/api/logs', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain'
        },
        body: line
      })))

      const failedResponse = responses.find((response) => !response.ok)
      if (!failedResponse) {
        setRawLogs('');
        toast.success(`${logLines.length} log${logLines.length === 1 ? '' : 's'} sent for processing!`);
      } else {
        const errorData = await failedResponse.json().catch(() => ({}))
        const successfulCount = responses.length - responses.filter((response) => !response.ok).length
        toast.error(`Processed ${successfulCount}/${logLines.length} logs. ${errorData.error || 'Some logs failed.'}`)
      }
    } catch (err) {
      console.error(err);
      toast.error('An error occurred while sending logs.');
    } finally {
      setIsProcessing(false);
    }
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>
      {mobileOpen && <div className="fixed inset-0 z-50 flex md:hidden"><div className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} /><div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div></div>}
      <main className="md:pl-60">
        <div className="mx-auto max-w-[1400px] px-5 py-5 sm:px-8 sm:py-7">
          <header className="mb-7 flex items-start justify-between gap-4">
            <div className="flex items-start gap-3"><Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu"><Menu /></Button><div><p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Overview</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">Dashboard</h1><p className="mt-1 text-sm text-muted-foreground">Overview of your GoLogGo workspace</p></div></div>
            <div className="flex items-center gap-2"><ThemeToggle /></div>
          </header>

          <section aria-label="Log metrics" className="grid grid-cols-2 gap-3 lg:grid-cols-4">
            {[['Total Requests', stats?.total ?? '—', Database], ['Completed', stats?.completed ?? '—', BarChart3], ['Waiting Parser', stats?.waitingParser ?? '—', CircleHelp], ['Failed', stats?.failed ?? '—', CircleHelp]].map(([label, value, Icon]) => <Card key={label} className="gap-3 py-4"><CardContent className="px-4"><div className="flex items-center justify-between"><p className="text-xs text-muted-foreground">{label}</p><Icon className="size-4 text-muted-foreground/70" /></div><div className="mt-2 flex items-end justify-between gap-2"><p className="font-mono text-xl font-medium tracking-tight">{value}</p></div></CardContent></Card>)}
          </section>

          <Card className="mt-5">
            <CardHeader className="border-b border-border px-5 py-4">
              <div>
                <CardTitle className="text-sm font-medium">Add Logs</CardTitle>
                <p className="mt-1 text-xs text-muted-foreground">Paste your logs below to process them.</p>
              </div>
            </CardHeader>
            <CardContent className="px-5 py-6">
              <textarea
                className="w-full min-h-[160px] rounded-lg border border-input bg-transparent px-3 py-2 text-sm transition-colors outline-none placeholder:text-muted-foreground/60 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-input/30 font-mono resize-y"
                placeholder={`Paste your log entries here...\n\n2024-09-09 10:15:23 [INFO] user-service - User registered\n2024-09-09 10:15:24 [ERROR] database - Connection timeout\n2024-09-09 10:15:25 [WARN] nginx - Invalid request`}
                value={rawLogs}
                onChange={(e) => setRawLogs(e.target.value)}
              />
              <div className="mt-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <p className="text-xs text-muted-foreground">Supports plain text logs. Each line will be processed separately.</p>
                <div className="flex items-center gap-2 self-end sm:self-auto">
                  <Button variant="outline" size="sm" onClick={handleClearLogs} disabled={isProcessing}>Clear</Button>
                  <Button size="sm" onClick={handleProcessLogs} disabled={isProcessing}>
                    {isProcessing ? 'Processing...' : 'Process Logs'}
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="mt-5"><CardHeader className="border-b border-border px-5 py-4"><div><CardTitle className="text-sm font-medium">Log Parsing Overview</CardTitle><p className="mt-1 text-xs text-muted-foreground">Successfully processed logs by parser</p></div></CardHeader><CardContent className="px-5 py-6">{parserData.length > 0 ? <div className="flex flex-col gap-4">{parserData.map((item) => <div key={item.name} className="grid grid-cols-[100px_1fr_64px] items-center gap-3 text-sm sm:grid-cols-[140px_1fr_76px]"><span className="truncate text-muted-foreground">{item.name}</span><div className="h-2 overflow-hidden rounded-sm bg-accent"><div className="h-full rounded-sm bg-primary/75" style={{ width: `${(item.count / parserData[0].count) * 100}%` }} /></div><span className="text-right font-mono text-xs text-muted-foreground">{item.count.toLocaleString()}</span></div>)}</div> : <p className="text-sm text-muted-foreground">No parser activity yet.</p>}</CardContent></Card>

          <Card className="mt-5"><CardHeader className="flex-row items-center justify-between px-5 py-4"><div><CardTitle className="text-sm font-medium">Recent Logs</CardTitle><p className="mt-1 text-xs text-muted-foreground">Latest events received by GoLogGo</p></div><Button variant="ghost" size="sm" className="text-xs text-muted-foreground hover:text-foreground" onClick={() => router.push('/logs')}>View all logs <span aria-hidden="true">→</span></Button></CardHeader><CardContent className="p-0"><div className="overflow-x-auto">{recentLogs.length > 0 ? <Table><TableHeader><TableRow><TableHead>Time</TableHead><TableHead>Level</TableHead><TableHead>Service</TableHead><TableHead>Message</TableHead><TableHead>Status</TableHead></TableRow></TableHeader><TableBody>{recentLogs.map((log) => <TableRow key={log.id} tabIndex={0} className="cursor-pointer" onClick={() => router.push(`/logs/${log.id}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') router.push(`/logs/${log.id}`) }}><TableCell className="font-mono text-xs text-muted-foreground">{log.time ? new Date(log.time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }) : '—'}</TableCell><TableCell><Badge variant={log.level === 'ERROR' ? 'destructive' : 'secondary'}>{log.level}</Badge></TableCell><TableCell className="font-mono text-xs">{log.service}</TableCell><TableCell className="min-w-48 max-w-[560px] truncate text-sm">{log.message}</TableCell><TableCell><Badge variant="outline" className={log.status === 'failed' ? 'border-destructive/40 text-destructive' : 'border-emerald-500/30 text-emerald-400'}>{log.status}</Badge></TableCell></TableRow>)}</TableBody></Table> : <p className="px-5 py-8 text-sm text-muted-foreground">No recent logs yet.</p>}</div></CardContent></Card>
        </div>
      </main>
    </div>
  )
}
