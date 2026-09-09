'use client'

import { useEffect, useMemo, useState } from 'react'
import { Sidebar } from '@/components/sidebar'
import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Check,
  Copy,
  Edit2,
  FileJson,
  Menu,
  Plus,
  RefreshCw,
  Search,
  X
} from 'lucide-react'
import { ThemeToggle } from '@/components/theme-toggle'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

function formatTimeAgo(dateInput) {
  if (!dateInput) return '—'
  const date = new Date(dateInput)
  if (isNaN(date.getTime())) return '—'
  const now = new Date()
  const diffSeconds = Math.floor((now - date) / 1000)

  if (diffSeconds < 60) return 'Just now'
  if (diffSeconds < 3600) return `${Math.floor(diffSeconds / 60)} min ago`
  if (diffSeconds < 86400) return `${Math.floor(diffSeconds / 3600)} hr ago`
  if (diffSeconds < 604800) return `${Math.floor(diffSeconds / 86400)} days ago`
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function formatDate(dateInput) {
  if (!dateInput) return '—'
  const date = new Date(dateInput)
  if (isNaN(date.getTime())) return '—'
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function SelectFilter({ label, value, onChange, options }) {
  return (
    <label className="relative flex min-w-32 items-center">
      <span className="sr-only">{label}</span>
      <select
        aria-label={label}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="h-9 w-full appearance-none rounded-md border border-input bg-background px-3 pr-8 text-xs text-foreground outline-none transition-colors focus:ring-1 focus:ring-ring"
      >
        {options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-2.5 size-3.5 text-muted-foreground" />
    </label>
  )
}

export function ParsersPage() {
  const [parsersList, setParsersList] = useState([])
  const [metrics, setMetrics] = useState({ totalParsers: 0, activeParsers: 0, totalLogsProcessed: 0 })
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState(null)

  // Filters & Search
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('All')
  const [format, setFormat] = useState('All Formats')
  const [mobileOpen, setMobileOpen] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  // Modals
  const [editingParser, setEditingParser] = useState(null)
  const [newName, setNewName] = useState('')
  const [newFormat, setNewFormat] = useState('')
  const [newStatus, setNewStatus] = useState('Active')
  const [savingEdit, setSavingEdit] = useState(false)

  const [inspectParser, setInspectParser] = useState(null)
  const [copiedHash, setCopiedHash] = useState(false)
  const [copiedPattern, setCopiedPattern] = useState(false)

  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createForm, setCreateForm] = useState({
    name: '',
    format: '',
    pattern: '',
    sampleLog: '',
    status: 'Active',
  })
  const [creating, setCreating] = useState(false)

  // Fetch parsers from API
  const fetchParsers = async (isManualRefresh = false) => {
    if (isManualRefresh) setRefreshing(true)
    else setLoading(true)
    setError(null)

    try {
      const res = await fetch('/api/parsers', { cache: 'no-store' })
      if (!res.ok) {
        throw new Error(`Failed to load parsers (${res.status})`)
      }
      const data = await res.json()
      const list = data.parsers || []
      setParsersList(list)

      if (data.metrics) {
        setMetrics(data.metrics)
      } else {
        const active = list.filter((p) => (p.status || 'Active').toLowerCase() === 'active').length
        const totalLogs = list.reduce((acc, p) => acc + (p.logsProcessed || 0), 0)
        setMetrics({
          totalParsers: list.length,
          activeParsers: active,
          totalLogsProcessed: totalLogs,
        })
      }
    } catch (err) {
      console.error('Error fetching parsers:', err)
      setError(err.message || 'Failed to connect to backend server')
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchParsers()
  }, [])

  // Available format filters from actual dataset
  const formatOptions = useMemo(() => {
    const formats = new Set(
      parsersList.map((p) => p.format || 'Custom').filter(Boolean)
    )
    return ['All Formats', ...Array.from(formats)]
  }, [parsersList])

  // Filtered parsers list
  const filtered = useMemo(() => {
    setCurrentPage(1)
    return parsersList.filter((parser) => {
      const name = parser.name || ''
      const fmt = parser.format || 'Custom'
      const vendor = parser.vendor || ''
      const product = parser.product || ''
      const hash = parser.hash || ''
      const pattern = parser.pattern || ''
      const sample = parser.sampleLog || ''
      const parserStatus = parser.status || 'Active'
      const fields = (parser.extractedFields || Object.keys(parser.mapping || {})).join(' ')

      const matchesSearch =
        `${name} ${fmt} ${vendor} ${product} ${hash} ${pattern} ${sample} ${fields}`
          .toLowerCase()
          .includes(search.toLowerCase())
      const matchesStatus = status === 'All' || parserStatus.toLowerCase() === status.toLowerCase()
      const matchesFormat = format === 'All Formats' || fmt.toLowerCase() === format.toLowerCase()

      return matchesSearch && matchesStatus && matchesFormat
    })
  }, [parsersList, search, status, format])

  const totalPages = Math.ceil(filtered.length / pageSize)
  const paginated = filtered.slice((currentPage - 1) * pageSize, currentPage * pageSize)

  const clear = () => {
    setSearch('')
    setStatus('All')
    setFormat('All Formats')
  }

  // Handle Edit / Rename
  const openEditModal = (p, e) => {
    e?.stopPropagation()
    setEditingParser(p)
    setNewName(p.name || '')
    setNewFormat(p.format || 'Custom')
    setNewStatus(p.status || 'Active')
  }

  const handleSaveEdit = async (e) => {
    e.preventDefault()
    if (!editingParser) return
    setSavingEdit(true)

    const parserId = editingParser.id || editingParser.hash
    try {
      const res = await fetch(`/api/parsers/${parserId}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newName.trim() || editingParser.name,
          format: newFormat.trim() || editingParser.format,
          status: newStatus,
        }),
      })

      if (!res.ok) {
        const errorData = await res.json().catch(() => ({}))
        throw new Error(errorData.error || 'Failed to update parser')
      }

      setParsersList((prev) =>
        prev.map((p) => {
          if ((p.id && p.id === editingParser.id) || (p.hash && p.hash === editingParser.hash)) {
            return {
              ...p,
              name: newName.trim() || p.name,
              format: newFormat.trim() || p.format,
              status: newStatus,
              updatedAt: new Date().toISOString(),
            }
          }
          return p
        })
      )
      if (
        inspectParser &&
        ((inspectParser.id && inspectParser.id === editingParser.id) ||
          inspectParser.hash === editingParser.hash)
      ) {
        setInspectParser((prev) => ({
          ...prev,
          name: newName.trim() || prev.name,
          format: newFormat.trim() || prev.format,
          status: newStatus,
        }))
      }
      setEditingParser(null)
    } catch (err) {
      toast.error(`Error updating parser: ${err.message}`)
    } finally {
      setSavingEdit(false)
    }
  }

  // Handle Create Parser
  const handleCreateParser = async (e) => {
    e.preventDefault()
    if (!createForm.pattern) {
      toast.error('Regex pattern is required')
      return
    }
    setCreating(true)

    try {
      const res = await fetch('/api/parsers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: createForm.name.trim() || undefined,
          format: createForm.format.trim() || 'Custom',
          pattern: createForm.pattern.trim(),
          sampleLog: createForm.sampleLog.trim() || undefined,
          status: createForm.status || 'Active',
        }),
      })

      if (!res.ok) {
        const errorData = await res.json().catch(() => ({}))
        throw new Error(errorData.error || 'Failed to create parser')
      }

      setCreateModalOpen(false)
      setCreateForm({ name: '', format: '', pattern: '', sampleLog: '', status: 'Active' })
      fetchParsers(true)
    } catch (err) {
      toast.error(`Error creating parser: ${err.message}`)
    } finally {
      setCreating(false)
    }
  }

  const copyToClipboard = (text, type = 'hash') => {
    navigator.clipboard.writeText(text)
    if (type === 'hash') {
      setCopiedHash(true)
      setTimeout(() => setCopiedHash(false), 2000)
    } else {
      setCopiedPattern(true)
      setTimeout(() => setCopiedPattern(false), 2000)
    }
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Sidebar Desktop */}
      <div className="fixed inset-y-0 left-0 hidden md:flex">
        <Sidebar />
      </div>

      {/* Sidebar Mobile */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 flex md:hidden">
          <button
            className="absolute inset-0 bg-background/80 backdrop-blur-sm"
            onClick={() => setMobileOpen(false)}
            aria-label="Close navigation"
          />
          <div className="relative">
            <Sidebar onClose={() => setMobileOpen(false)} />
          </div>
        </div>
      )}

      {/* Main Content Area */}
      <main className="md:pl-60">
        <div className="mx-auto max-w-[1500px] px-5 py-5 sm:px-8 sm:py-7">
          {/* Header */}
          <header className="mb-7 flex items-start justify-between gap-4">
            <div className="flex items-start gap-3">
              <Button
                variant="ghost"
                size="icon"
                className="mt-0.5 size-9 md:hidden"
                onClick={() => setMobileOpen(true)}
                aria-label="Open menu"
              >
                <Menu />
              </Button>
              <div>
                <p className="font-mono text-xs uppercase tracking-widest text-muted-foreground">
                  Workspace
                </p>
                <h1 className="mt-1 text-2xl font-semibold tracking-tight">Parsers</h1>
                <p className="mt-1 text-sm text-muted-foreground">
                  Manage and inspect active log parsers and extraction schemas
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <ThemeToggle />
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                onClick={() => fetchParsers(true)}
                disabled={loading || refreshing}
                aria-label="Refresh parsers"
              >
                <RefreshCw className={`size-3.5 ${refreshing ? 'animate-spin' : ''}`} />
              </Button>
              <Button
                size="sm"
                className="gap-2"
                onClick={() => setCreateModalOpen(true)}
              >
                <Plus className="size-3.5" />
                <span>Create Parser</span>
              </Button>
            </div>
          </header>

          {/* Metrics Section */}
          <section
            aria-label="Parser metrics"
            className="mb-5 grid grid-cols-3 divide-x divide-border rounded-lg border border-border bg-card/40"
          >
            <div className="px-4 py-4">
              <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Total Parsers</p>
              <p className="mt-1 font-mono text-xl">{loading ? '—' : metrics.totalParsers || parsersList.length}</p>
            </div>
            <div className="px-4 py-4">
              <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Active Parsers</p>
              <p className="mt-1 font-mono text-xl text-emerald-400">{loading ? '—' : metrics.activeParsers || 0}</p>
            </div>
            <div className="px-4 py-4">
              <p className="text-[11px] uppercase tracking-wider text-muted-foreground">Logs Processed</p>
              <p className="mt-1 font-mono text-xl">
                {loading ? '—' : Number(metrics.totalLogsProcessed || 0).toLocaleString()}
              </p>
            </div>
          </section>

          {/* Error Message */}
          {error && (
            <div className="mb-5 rounded-lg border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-400">
              <div className="flex items-center justify-between">
                <span>{error}</span>
                <Button size="xs" variant="outline" onClick={() => fetchParsers(true)}>
                  Retry
                </Button>
              </div>
            </div>
          )}

          {/* Filters Bar */}
          <section
            aria-label="Parser filters"
            className="mb-5 flex flex-col gap-3 rounded-lg border border-border bg-card/40 p-3"
          >
            <div className="relative">
              <Search className="absolute left-3 top-2.5 size-4 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Search parsers by name, format, or fields..."
                className="h-9 border-input bg-background pl-9 text-sm"
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <SelectFilter
                label="Status"
                value={status}
                onChange={setStatus}
                options={['All', 'Active', 'Inactive']}
              />
              <SelectFilter
                label="Format"
                value={format}
                onChange={setFormat}
                options={formatOptions}
              />
              {(search || status !== 'All' || format !== 'All Formats') && (
                <button
                  onClick={clear}
                  className="px-2 text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
                >
                  Clear filters
                </button>
              )}
            </div>
          </section>

          {/* Clean Parsers Table */}
          <section className="overflow-hidden rounded-lg border border-border bg-card/20">
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <div>
                <h2 className="text-sm font-medium">All Parsers</h2>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {parsersList.length} total parsers in database
                </p>
              </div>
              <span className="font-mono text-xs text-muted-foreground">MongoDB Synced</span>
            </div>

            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow className="hover:bg-transparent">
                    <TableHead className="w-48 pl-4">Parser Name</TableHead>
                    <TableHead className="w-28">Format</TableHead>
                    <TableHead className="min-w-48">Extracted Fields</TableHead>
                    <TableHead className="min-w-64">Sample Log</TableHead>
                    <TableHead className="w-28">Logs</TableHead>
                    <TableHead className="w-24">Status</TableHead>
                    <TableHead className="w-28">Last Used</TableHead>
                    <TableHead className="w-28 pr-4 text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loading ? (
                    <TableRow>
                      <TableCell colSpan={8} className="h-32 text-center text-xs text-muted-foreground">
                        Loading parsers...
                      </TableCell>
                    </TableRow>
                  ) : paginated.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={8} className="h-32 text-center text-xs text-muted-foreground">
                        No parsers found
                      </TableCell>
                    </TableRow>
                  ) : (
                    paginated.map((parserItem) => {
                      const parserName = parserItem.name || 'Unnamed Parser'
                      const parserFormat = parserItem.format || 'Custom'
                      const parserStatus = parserItem.status || 'Active'
                      const isActive = parserStatus.toLowerCase() === 'active'
                      const fields = parserItem.extractedFields || Object.keys(parserItem.mapping || {})

                      return (
                        <TableRow
                          key={parserItem.id || parserItem.hash}
                          className="cursor-pointer transition-colors hover:bg-accent/50 group"
                          onClick={() => setInspectParser(parserItem)}
                        >
                          {/* Parser Name */}
                          <TableCell className="pl-4 font-mono text-xs">
                            <div className="flex items-center gap-1.5">
                              <span className="font-medium text-foreground">{parserName}</span>
                              <button
                                onClick={(e) => openEditModal(parserItem, e)}
                                className="opacity-0 group-hover:opacity-100 p-0.5 text-muted-foreground hover:text-foreground transition-opacity"
                                title="Rename Parser"
                              >
                                <Edit2 className="size-3" />
                              </button>
                            </div>
                            <span className="text-[10px] text-muted-foreground block truncate max-w-36">
                              {parserItem.hash ? `${parserItem.hash.slice(0, 12)}...` : ''}
                            </span>
                          </TableCell>

                          {/* Format */}
                          <TableCell>
                            <Badge variant="outline" className="font-mono text-[10px]">
                              {parserFormat}
                            </Badge>
                          </TableCell>

                          {/* Extracted Fields */}
                          <TableCell>
                            <div className="flex flex-wrap items-center gap-1 max-w-[220px]">
                              {fields.length === 0 ? (
                                <span className="text-xs text-muted-foreground">—</span>
                              ) : (
                                <>
                                  {fields.slice(0, 3).map((f) => (
                                    <span
                                      key={f}
                                      className="rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground"
                                    >
                                      {f}
                                    </span>
                                  ))}
                                  {fields.length > 3 && (
                                    <span className="text-[10px] text-muted-foreground font-mono">
                                      +{fields.length - 3}
                                    </span>
                                  )}
                                </>
                              )}
                            </div>
                          </TableCell>

                          {/* Sample Log Preview */}
                          <TableCell className="max-w-xs font-mono text-xs text-muted-foreground truncate">
                            {parserItem.sampleLog || parserItem.pattern || '—'}
                          </TableCell>

                          {/* Logs Count */}
                          <TableCell className="font-mono text-xs">
                            {Number(parserItem.logsProcessed || 0).toLocaleString()}
                          </TableCell>

                          {/* Status */}
                          <TableCell>
                            <span
                              className={`inline-flex items-center gap-1.5 text-xs ${
                                isActive ? 'text-emerald-400' : 'text-muted-foreground'
                              }`}
                            >
                              <span
                                className={`size-1.5 rounded-full ${
                                  isActive ? 'bg-emerald-400' : 'bg-muted-foreground'
                                }`}
                              />
                              {parserStatus}
                            </span>
                          </TableCell>

                          {/* Last Used */}
                          <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                            {formatTimeAgo(parserItem.lastUsed || parserItem.updatedAt)}
                          </TableCell>

                          {/* Actions */}
                          <TableCell className="pr-4 text-right">
                            <div className="flex items-center justify-end gap-1">
                              <Button
                                variant="ghost"
                                size="sm"
                                className="h-7 px-2 text-xs text-muted-foreground hover:text-foreground"
                                onClick={(e) => openEditModal(parserItem, e)}
                              >
                                Rename
                              </Button>
                              <span className="text-muted-foreground text-xs">›</span>
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })
                  )}
                </TableBody>
              </Table>
            </div>
          </section>

          {/* Pagination Footer */}
          <div className="flex flex-col gap-4 py-4 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
            <span>
              Showing {filtered.length === 0 ? 0 : (currentPage - 1) * pageSize + 1}–
              {Math.min(currentPage * pageSize, filtered.length)} of {filtered.length} parsers
            </span>
            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                className="h-8 gap-1"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              >
                <ChevronLeft className="size-3.5" />
                Previous
              </Button>
              <div className="flex items-center gap-1 mx-2 text-sm font-medium">
                Page {currentPage} of {totalPages || 1}
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-8 gap-1"
                disabled={currentPage === totalPages || totalPages === 0}
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              >
                Next
                <ChevronRight className="size-3.5" />
              </Button>
              <select
                aria-label="Parsers per page"
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value))
                  setCurrentPage(1)
                }}
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

      {/* RENAME / EDIT PARSER MODAL */}
      {editingParser && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4">
          <div className="relative w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-lg">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h2 className="text-base font-semibold">Rename Parser</h2>
              <button
                onClick={() => setEditingParser(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="size-4" />
              </button>
            </div>

            <form onSubmit={handleSaveEdit} className="mt-4 space-y-4 text-xs">
              <div>
                <label className="block font-medium mb-1">Parser Name</label>
                <Input
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="e.g. NginxParser"
                  className="h-8 text-xs"
                  autoFocus
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-medium mb-1">Format</label>
                  <Input
                    value={newFormat}
                    onChange={(e) => setNewFormat(e.target.value)}
                    placeholder="e.g. Nginx"
                    className="h-8 text-xs"
                  />
                </div>

                <div>
                  <label className="block font-medium mb-1">Status</label>
                  <select
                    value={newStatus}
                    onChange={(e) => setNewStatus(e.target.value)}
                    className="h-8 w-full rounded-md border border-input bg-background px-2 text-xs text-foreground outline-none"
                  >
                    <option value="Active">Active</option>
                    <option value="Inactive">Inactive</option>
                  </select>
                </div>
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t border-border">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setEditingParser(null)}
                >
                  Cancel
                </Button>
                <Button type="submit" size="sm" disabled={savingEdit}>
                  Save
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* INSPECT DETAILS MODAL */}
      {inspectParser && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4">
          <div className="relative w-full max-w-2xl rounded-lg border border-border bg-card p-6 shadow-xl max-h-[85vh] overflow-y-auto">
            <div className="flex items-start justify-between border-b border-border pb-4">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-base font-semibold">{inspectParser.name || 'Parser Details'}</h2>
                  <Badge variant="outline" className="text-[10px]">
                    {inspectParser.format || 'Custom'}
                  </Badge>
                </div>
                <p className="text-xs text-muted-foreground mt-0.5">
                  Fingerprint Hash: <span className="font-mono">{inspectParser.hash || '—'}</span>
                </p>
              </div>
              <button
                onClick={() => setInspectParser(null)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="size-4" />
              </button>
            </div>

            <div className="mt-4 space-y-4 text-xs">
              {/* Sample Log */}
              {inspectParser.sampleLog && (
                <div>
                  <span className="font-medium text-muted-foreground block mb-1">Sample Raw Log</span>
                  <div className="rounded-md border border-border bg-muted/30 p-2.5 font-mono text-xs text-foreground whitespace-pre-wrap break-all">
                    {inspectParser.sampleLog}
                  </div>
                </div>
              )}

              {/* Regex Pattern */}
              <div>
                <div className="flex items-center justify-between mb-1">
                  <span className="font-medium text-muted-foreground">Synthesized Regex Pattern</span>
                  <button
                    onClick={() => copyToClipboard(inspectParser.pattern, 'pat')}
                    className="text-[11px] text-muted-foreground hover:text-foreground flex items-center gap-1"
                  >
                    {copiedPattern ? <Check className="size-3 text-emerald-400" /> : <Copy className="size-3" />}
                    <span>{copiedPattern ? 'Copied' : 'Copy'}</span>
                  </button>
                </div>
                <div className="rounded-md border border-border bg-muted/30 p-2.5 font-mono text-xs text-foreground whitespace-pre-wrap break-all">
                  {inspectParser.pattern || 'No regex pattern defined'}
                </div>
              </div>

              {/* Extracted Fields */}
              <div>
                <span className="font-medium text-muted-foreground block mb-1">Extracted Fields</span>
                <div className="flex flex-wrap gap-1.5">
                  {(inspectParser.extractedFields || Object.keys(inspectParser.mapping || {})).map((f) => (
                    <span
                      key={f}
                      className="rounded border border-border bg-muted/40 px-2 py-1 font-mono text-xs"
                    >
                      {f}
                    </span>
                  ))}
                </div>
              </div>

              {/* Field Mappings */}
              {inspectParser.mapping && Object.keys(inspectParser.mapping).length > 0 && (
                <div>
                  <span className="font-medium text-muted-foreground block mb-1">Field Normalization</span>
                  <div className="rounded-md border border-border overflow-hidden">
                    <Table>
                      <TableHeader>
                        <TableRow className="hover:bg-transparent">
                          <TableHead className="text-xs">Extracted Key</TableHead>
                          <TableHead className="text-xs">Normalized Field</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {Object.entries(inspectParser.mapping).map(([k, v]) => (
                          <TableRow key={k}>
                            <TableCell className="font-mono text-xs">{k}</TableCell>
                            <TableCell className="font-mono text-xs text-muted-foreground">{v}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                </div>
              )}

              {/* Metadata */}
              <div className="grid grid-cols-2 gap-2 pt-2 border-t border-border text-muted-foreground text-xs">
                <div>
                  Logs Processed: <span className="text-foreground font-mono">{inspectParser.logsProcessed || 0}</span>
                </div>
                <div>
                  Created: <span className="text-foreground">{formatDate(inspectParser.createdAt)}</span>
                </div>
              </div>
            </div>

            <div className="mt-5 flex justify-end border-t border-border pt-3">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setInspectParser(null)}
              >
                Close
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* CREATE PARSER MODAL */}
      {createModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm p-4">
          <div className="relative w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-lg">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h2 className="text-base font-semibold">Create Parser</h2>
              <button
                onClick={() => setCreateModalOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="size-4" />
              </button>
            </div>

            <form onSubmit={handleCreateParser} className="mt-4 space-y-3 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block font-medium mb-1">Parser Name</label>
                  <Input
                    value={createForm.name}
                    onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                    placeholder="e.g. AuthParser"
                    className="h-8 text-xs"
                    required
                  />
                </div>
                <div>
                  <label className="block font-medium mb-1">Format</label>
                  <Input
                    value={createForm.format}
                    onChange={(e) => setCreateForm({ ...createForm, format: e.target.value })}
                    placeholder="e.g. Nginx"
                    className="h-8 text-xs"
                  />
                </div>
              </div>

              <div>
                <label className="block font-medium mb-1">Regex Pattern *</label>
                <textarea
                  value={createForm.pattern}
                  onChange={(e) => setCreateForm({ ...createForm, pattern: e.target.value })}
                  placeholder="(?P<timestamp>\S+ \S+) (?P<severity>\w+) (?P<service>[\w-]+) (?P<message>.*)"
                  rows={3}
                  className="w-full rounded-md border border-input bg-background p-2 font-mono text-xs text-foreground outline-none focus:ring-1 focus:ring-ring"
                  required
                />
              </div>

              <div>
                <label className="block font-medium mb-1">Sample Raw Log (Optional)</label>
                <textarea
                  value={createForm.sampleLog}
                  onChange={(e) => setCreateForm({ ...createForm, sampleLog: e.target.value })}
                  placeholder="2026-09-08 20:15:32 ERROR auth-service User login failed"
                  rows={2}
                  className="w-full rounded-md border border-input bg-background p-2 font-mono text-xs text-foreground outline-none focus:ring-1 focus:ring-ring"
                />
              </div>

              <div className="flex justify-end gap-2 pt-3 border-t border-border">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setCreateModalOpen(false)}
                >
                  Cancel
                </Button>
                <Button type="submit" size="sm" disabled={creating}>
                  Create
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
