'use client'

import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'

export function ThemeToggle() {
  const [isLight, setIsLight] = useState(false)

  useEffect(() => {
    const saved = window.localStorage.getItem('log-ai-theme')
    const light = saved === 'light'
    setIsLight(light)
    document.documentElement.classList.toggle('light', light)
    document.documentElement.classList.toggle('dark', !light)
  }, [])

  const toggleTheme = () => {
    const nextIsLight = !isLight
    setIsLight(nextIsLight)
    window.localStorage.setItem('log-ai-theme', nextIsLight ? 'light' : 'dark')
    document.documentElement.classList.toggle('light', nextIsLight)
    document.documentElement.classList.toggle('dark', !nextIsLight)
  }

  return (
    <button
      type="button"
      onClick={toggleTheme}
      className="inline-flex h-8 items-center gap-1 rounded-md border border-border bg-card px-2 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
      aria-label={`Switch to ${isLight ? 'dark' : 'light'} theme`}
      title={`Switch to ${isLight ? 'dark' : 'light'} theme`}
    >
      {isLight ? <Sun aria-hidden="true" /> : <Moon aria-hidden="true" />}
      <span className="hidden sm:inline">{isLight ? 'Light' : 'Dark'}</span>
    </button>
  )
}
