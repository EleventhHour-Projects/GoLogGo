'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import { useRouter } from 'next/navigation'
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Menu,
  RefreshCw,
  Search,
} from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

// ─── Status helpers ──────────────────────────────────────────────────────────

const STATUS_META = {
  pending:        { label: 'Pending',        color: 'text-amber-400',   dot: 'bg-amber-400',   pulse: true },
  processing:     { label: 'Processing',     color: 'text-blue-400',    dot: 'bg-blue-400',    pulse: true },
  waiting_parser: { label: 'Waiting Parser', color: 'text-violet-400',  dot: 'bg-violet-400',  pulse: true },
  completed:      { label: 'Completed',      color: 'text-emerald-400', dot: 'bg-emerald-400', pulse: false },
  failed:         { label: 'Failed',         color: 'text-destructive', dot: 'bg-destructive', pulse: false },
}

function StatusPill({ status }) {
  const meta = STATUS_META[status] ?? { label: status, color: 'text-muted-foreground', dot: 'bg-muted-foreground', pulse: false }
  return (
    <span className={`inline-flex items-center gap-1.5 text-xs font-medium ${meta.color}`}>
      <span className={`size-1.5 rounded-full ${meta.dot} ${meta.pulse ? 'animate-pulse' : ''}`} />
      {meta.label}
    </span>
  )
}

// ─── Filter select ────────────────────────────────────────────────────────────

function FilterSelect({ label, value, onChange, options }) {
  return (
    <label className="relative flex min-w-[128px] flex-1 items-center">
      <span className="sr-only">{label}</span>
      <select
        aria-label={label}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-9 w-full appearance-none rounded-md border border-input bg-background px-3 pr-8 text-xs text-foreground outline-none transition-colors focus:ring-1 focus:ring-ring"
      >
        {options.map((opt) => <option key={opt.value} value={opt.value}>{opt.label}</option>)}
      </select>
      <ChevronDown className="pointer-events-none absolute right-2.5 size-3.5 text-muted-foreground" />
    </label>
  )
}

// ─── Stat card ────────────────────────────────────────────────────────────────

function StatCard({ label, value, colorClass = 'text-foreground' }) {
  return (
    <div className="flex flex-col gap-1 rounded-lg border border-border bg-card/40 px-4 py-3">
      <p className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className={`font-mono text-xl font-medium ${colorClass}`}>{value ?? '—'}</p>
    </div>
  )
}

// ─── Main component ───────────────────────────────────────────────────────────

export function LogsExplorer() {
  const router = useRouter()
  const [mobileOpen, setMobileOpen] = useState(false)

  // Fetch state
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState(null)
  const [stats, setStats] = useState(null)
  const [requests, setRequests] = useState([])
  const [totalCount, setTotalCount] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [isAuthenticated, setIsAuthenticated] = useState(true)

  // Filter / pagination state
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const statusOptions = [
    { value: 'all',            label: 'All Statuses' },
    { value: 'pending',        label: 'Pending' },
    { value: 'processing',     label: 'Processing' },
    { value: 'waiting_parser', label: 'Waiting Parser' },
    { value: 'completed',      label: 'Completed' },
    { value: 'failed',         label: 'Failed' },
  ]

  // ── Fetch data from /api/logs ──────────────────────────────────────────────

  const fetchLogs = useCallback(async (opts = {}) => {
    const isRefresh = opts.refresh ?? false
    if (isRefresh) setRefreshing(true)
    else setLoading(true)
    setError(null)

    try {
      const params = new URLSearchParams({
        page:   String(opts.page   ?? currentPage),
        limit:  String(opts.limit  ?? pageSize),
        status: opts.status ?? statusFilter,
      })

      const res = await fetch(`/api/logs?${params}`)

      if (res.status === 401) {
        setIsAuthenticated(false)
        return
      }

      if (!res.ok) throw new Error('Failed to fetch logs')

      const data = await res.json()
      setStats(data.stats)
      setRequests(data.requests ?? [])
      setTotalCount(data.total ?? 0)
      setTotalPages(data.totalPages ?? 1)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [currentPage, pageSize, statusFilter, router])

  // Initial load
  useEffect(() => { fetchLogs() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  // Re-fetch on filter/page change (skip on mount, handled above)
  useEffect(() => {
    fetchLogs({ page: currentPage, limit: pageSize, status: statusFilter })
  }, [currentPage, pageSize, statusFilter]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleRefresh = () => {
    fetchLogs({ refresh: true, page: currentPage, limit: pageSize, status: statusFilter })
  }

  const clearFilters = () => {
    setSearch('')
    setStatusFilter('all')
    setCurrentPage(1)
  }

  // Client-side search on the current page results (search within payload text)
  const filtered = useMemo(() => {
    if (!search.trim()) return requests
    const q = search.toLowerCase()
    return requests.filter((r) => {
      const payloadMatch = r.payload?.toLowerCase().includes(q)
      const idMatch = r.id?.toLowerCase().includes(q)
      const statusMatch = r.status?.toLowerCase().includes(q)
      const logsMatch = r.logs?.some?.((l) => JSON.stringify(l.normalizedLog).toLowerCase().includes(q))
      return payloadMatch || idMatch || statusMatch || logsMatch
    })
  }, [requests, search])

  // ── Format helpers ─────────────────────────────────────────────────────────

  const formatTime = (iso) => {
    if (!iso) return '—'
    try {
      return new Date(iso).toLocaleString('en-IN', {
        dateStyle: 'short', timeStyle: 'medium', hour12: false,
      })
    } catch { return iso }
  }

  const trimPayload = (payload, len = 80) => {
    if (!payload) return '—'
    const first = payload.split('\n')[0].trim()
    return first.length > len ? first.slice(0, len) + '…' : first
  }

  // ── Render ─────────────────────────────────────────────────────────────────

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-background text-foreground">
        <div className="fixed inset-y-0 left-0 hidden md:flex">
          <Sidebar />
        </div>

        {mobileOpen && (
          <div className="fixed inset-0 z-50 flex md:hidden">
            <button
              className="absolute inset-0 bg-background/80"
              onClick={() => setMobileOpen(false)}
              aria-label="Close navigation"
            />
            <div className="relative">
              <Sidebar onClose={() => setMobileOpen(false)} />
            </div>
          </div>
        )}

        <main className="md:pl-60">
          <div className="flex min-h-screen items-center justify-center px-5">
            <div className="w-full max-w-md text-center">
              <h1 className="text-xl font-semibold tracking-tight">
                Sign in to view your logs
              </h1>

              <p className="mt-2 text-sm leading-6 text-muted-foreground">
                Sign in to access your submitted logs, processing status,
                and detailed log information.
              </p>

              <Button
                className="mt-6"
                onClick={() => router.push('/login')}
              >
                Login
              </Button>
            </div>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 hidden md:flex"><Sidebar /></div>
      {mobileOpen && (
        <div className="fixed inset-0 z-50 flex md:hidden">
          <button className="absolute inset-0 bg-background/80" onClick={() => setMobileOpen(false)} aria-label="Close navigation" />
          <div className="relative"><Sidebar onClose={() => setMobileOpen(false)} /></div>
        </div>
      )}

      <main className="md:pl-60">
        <div className="mx-auto max-w-[1500px] px-5 py-5 sm:px-8 sm:py-7">

          {/* ── Header ── */}
          <header className="mb-7 flex items-start justify-between gap-4">
            <div className="flex items-start gap-3">
              <Button variant="ghost" size="icon" className="mt-0.5 size-9 md:hidden" onClick={() => setMobileOpen(true)} aria-label="Open menu">
                <Menu />
              </Button>
              <div>
                <p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">Workspace</p>
                <h1 className="mt-1 text-2xl font-semibold tracking-tight">Logs</h1>
                <p className="mt-1 text-sm text-muted-foreground">Browse and investigate your submitted log events</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <ThemeToggle />
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                onClick={handleRefresh}
                aria-label="Refresh logs"
                disabled={refreshing || loading}
              >
                <RefreshCw className={`size-3.5 ${refreshing ? 'animate-spin' : ''}`} />
              </Button>
            </div>
          </header>

          {/* ── Stats ── */}
          <section aria-label="Log request metrics" className="mb-5 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
            <StatCard label="Total"          value={stats?.total}         />
            <StatCard label="Pending"        value={stats?.pending}        colorClass="text-amber-400" />
            <StatCard label="Processing"     value={stats?.processing}     colorClass="text-blue-400" />
            <StatCard label="Waiting Parser" value={stats?.waitingParser}  colorClass="text-violet-400" />
            <StatCard label="Completed"      value={stats?.completed}      colorClass="text-emerald-400" />
            <StatCard label="Failed"         value={stats?.failed}         colorClass="text-destructive" />
          </section>

          {/* ── Filters ── */}
          <section aria-label="Log filters" className="mb-5 flex flex-col gap-3 rounded-lg border border-border bg-card/40 p-3">
            <div className="relative">
              <Search className="absolute left-3 top-2.5 size-4 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search by payload, ID, or log fields…"
                className="h-9 border-input bg-background pl-9 text-sm"
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <FilterSelect
                label="Status"
                value={statusFilter}
                onChange={(v) => { setStatusFilter(v); setCurrentPage(1) }}
                options={statusOptions}
              />
              <button
                onClick={clearFilters}
                className="px-2 text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
              >
                Clear filters
              </button>
            </div>
          </section>

          {/* ── Table ── */}
          <section className="overflow-hidden rounded-lg border border-border bg-card/20">
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <div>
                <h2 className="text-sm font-medium">Your Log Requests</h2>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {loading ? 'Loading…' : `${totalCount.toLocaleString()} request${totalCount !== 1 ? 's' : ''} found`}
                </p>
              </div>
              {stats?.lastActivityAt && (
                <span className="font-mono text-xs text-muted-foreground hidden sm:block">
                  Last activity: {formatTime(stats.lastActivityAt)}
                </span>
              )}
            </div>

            {error ? (
              <div className="flex items-center justify-center py-16 text-sm text-destructive">
                {error} —{' '}
                <button className="ml-1 underline" onClick={handleRefresh}>retry</button>
              </div>
            ) : loading ? (
              <div className="flex items-center justify-center py-16">
                <RefreshCw className="size-5 animate-spin text-muted-foreground" />
              </div>
            ) : filtered.length === 0 ? (
              <div className="flex flex-col items-center justify-center gap-2 py-16 text-center text-sm text-muted-foreground">
                <p>No logs found.</p>
                {(search || statusFilter !== 'all') && (
                  <button onClick={clearFilters} className="text-xs underline">Clear filters</button>
                )}
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="hover:bg-transparent">
                      <TableHead className="pl-4 w-40">Submitted At</TableHead>
                      <TableHead className="min-w-72">Payload Preview</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="w-20 text-center">Attempts</TableHead>
                      <TableHead className="w-28">Parsed Logs</TableHead>
                      <TableHead className="w-8 pr-4" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filtered.map((req) => (
                      <TableRow
                        key={req.id}
                        tabIndex={0}
                        className="cursor-pointer transition-colors hover:bg-accent/50"
                        onClick={() => router.push(`/logs/${req.id}`)}
                        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') router.push(`/logs/${req.id}`) }}
                      >
                        <TableCell className="pl-4 font-mono text-xs text-muted-foreground">
                          {formatTime(req.createdAt)}
                        </TableCell>
                        <TableCell className="text-xs text-foreground">
                          <span className="font-mono">{trimPayload(req.payload)}</span>
                        </TableCell>
                        <TableCell>
                          <StatusPill status={req.status} />
                        </TableCell>
                        <TableCell className="text-center font-mono text-xs text-muted-foreground">
                          {req.attempts}
                        </TableCell>
                        <TableCell>
                          {req.logs != null ? (
                            <Badge variant="outline" className="border-emerald-500/30 text-emerald-400 font-mono text-[10px]">
                              {req.logs.length} log{req.logs.length !== 1 ? 's' : ''}
                            </Badge>
                          ) : (
                            <span className="text-xs text-muted-foreground">—</span>
                          )}
                        </TableCell>
                        <TableCell className="pr-4 text-right text-muted-foreground">›</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </section>

          {/* ── Pagination (kept as-is from original design) ── */}
          <div className="flex flex-col gap-4 py-4 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <span>
              {loading || filtered.length === 0
                ? 'No results'
                : `Showing ${(currentPage - 1) * pageSize + 1}–${Math.min(currentPage * pageSize, totalCount)} of ${totalCount} requests`}
            </span>
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                className="h-8 gap-1"
                disabled={currentPage === 1 || loading}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              >
                <ChevronLeft className="size-3.5" />Previous
              </Button>
              <div className="flex items-center gap-1 mx-2 text-sm font-medium">
                Page {currentPage} of {totalPages || 1}
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-8 gap-1"
                disabled={currentPage >= totalPages || loading}
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              >
                Next<ChevronRight className="size-3.5" />
              </Button>
              <select
                aria-label="Requests per page"
                value={pageSize}
                onChange={(e) => { setPageSize(Number(e.target.value)); setCurrentPage(1) }}
                className="ml-2 h-8 rounded-md border border-input bg-background px-2 text-xs"
              >
                <option value={10}>10 per page</option>
                <option value={25}>25 per page</option>
                <option value={50}>50 per page</option>
              </select>
            </div>
          </div>

        </div>
      </main>
    </div>
  )
}

