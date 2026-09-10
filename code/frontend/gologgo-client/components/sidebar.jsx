'use client'

import { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { FileJson, LayoutDashboard, ListFilter, X } from 'lucide-react'
import { Button } from '@/components/ui/button'

const navItems = [
  { label: 'Dashboard', icon: LayoutDashboard, href: '/' },
  { label: 'Logs', icon: ListFilter, href: '/logs' },
  { label: 'Parsers', icon: FileJson, href: '/parsers' },
]

export function Sidebar({ onClose }) {
  const router = useRouter()
  const pathname = usePathname()
  const [profile, setProfile] = useState(null)

  useEffect(() => {
    fetch('/api/auth/me')
      .then((response) => {
        if (response.status === 401) {
          window.location.href = '/login'
          return null
        }
        return response.ok ? response.json() : null
      })
      .then(setProfile)
      .catch(() => setProfile(null))
  }, [])

  const initials = profile?.name
    ?.split(' ')
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()

  return (
    <aside className="flex h-full w-60 shrink-0 flex-col border-r border-border bg-sidebar px-3 py-4">
      <div className="mb-8 flex items-center justify-between px-3">
        <div className="flex items-center gap-2.5">
          <img src="/logo.png" alt="GoLogGo logo" className="size-9 object-contain" />
          <span className="font-mono text-sm font-semibold tracking-tight">GoLogGo</span>
        </div>
        {onClose && (
          <Button variant="ghost" size="icon" className="size-8 md:hidden" onClick={onClose} aria-label="Close menu">
            <X />
          </Button>
        )}
      </div>
      <nav aria-label="Primary navigation" className="flex flex-col gap-1">
        {navItems.map(({ label, icon: Icon, href }) => {
          const isActive = pathname === href || (href !== '/' && pathname?.startsWith(href))
          return (
            <button
              key={label}
              onClick={() => {
                router.push(href)
                if (onClose) onClose()
              }}
              className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors ${
                isActive ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground'
              }`}
            >
              <Icon className="size-4" />
              {label}
            </button>
          )
        })}
      </nav>
      <div className="mt-auto flex flex-col gap-1">
        <div className="mt-3 flex items-center justify-between border-t border-border px-3 pt-4">
          <div className="flex items-center gap-3 min-w-0">
            <div className="flex size-8 shrink-0 items-center justify-center overflow-hidden rounded-full bg-accent font-mono text-xs text-muted-foreground">
              {profile?.picture ? <img src={profile.picture} alt="" className="size-full object-cover" /> : initials || '...'}</div>
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{profile?.name || 'Loading...'}</p>
              <p className="truncate text-xs text-muted-foreground">{profile?.email || ''}</p>
            </div>
          </div>
          <button
            onClick={async () => {
              await fetch('/api/auth/logout', { method: 'POST' })
              window.location.href = '/login'
            }}
            className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground"
            title="Log out"
            aria-label="Log out"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" x2="9" y1="12" y2="12" />
            </svg>
          </button>
        </div>
      </div>
    </aside>
  )
}
