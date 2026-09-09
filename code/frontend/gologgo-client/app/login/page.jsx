'use client'

import { useEffect, useRef, useState } from 'react'
import { ThemeToggle } from '@/components/theme-toggle'

export default function LoginPage() {
  const googleButtonRef = useRef(null)
  const [error, setError] = useState('')
  const [isSigningIn, setIsSigningIn] = useState(false)

  useEffect(() => {
    const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID

    if (!clientId) {
      setError('Google sign-in is not configured. Add NEXT_PUBLIC_GOOGLE_CLIENT_ID to .env.')
      return undefined
    }

    const handleCredential = async ({ credential }) => {
      setError('')
      setIsSigningIn(true)

      try {
        const response = await fetch('/api/auth', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ credential }),
        })

        if (!response.ok) {
          const data = await response.json().catch(() => ({}))
          throw new Error(data.error || 'Unable to sign in with Google.')
        }

        const destination = new URLSearchParams(window.location.search).get('from') || '/'
        window.location.href = destination.startsWith('/') ? destination : '/'
      } catch (signInError) {
        setError(signInError.message)
        setIsSigningIn(false)
      }
    }

    let retryTimer
    let attempts = 0

    const renderButton = () => {
      if (!window.google?.accounts?.id || !googleButtonRef.current) {
        if (attempts < 50) {
          attempts += 1
          retryTimer = window.setTimeout(renderButton, 100)
        }
        return
      }

      if (googleButtonRef.current.childElementCount > 0) return
      window.google.accounts.id.initialize({ client_id: clientId, callback: handleCredential })
      window.google.accounts.id.renderButton(googleButtonRef.current, {
        theme: 'outline',
        size: 'large',
        width: 336,
        text: 'signin_with',
      })
    }

    const existingScript = document.querySelector('script[data-google-identity]')
    if (existingScript) {
      renderButton()
      return () => window.clearTimeout(retryTimer)
    }

    const script = document.createElement('script')
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.dataset.googleIdentity = 'true'
    script.onload = renderButton
    document.head.appendChild(script)

    return () => {
      script.onload = null
      window.clearTimeout(retryTimer)
    }
  }, [])

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground p-4">
      <div className="absolute top-4 right-4"><ThemeToggle /></div>
      
      <div className="w-full max-w-[400px] flex flex-col items-center">
        <div className="mb-8 flex items-center gap-3">
          <img src="/logo.png" alt="GoLogGo logo" className="size-20 object-contain" />
          <span className="font-mono text-3xl font-bold tracking-tight">GoLogGo</span>
        </div>

        <div className="w-full rounded-xl border border-border bg-card p-8 shadow-sm">
          <div className="mb-6 flex flex-col items-center text-center">
            <h1 className="text-2xl font-semibold tracking-tight">Sign in to GoLogGo</h1>
            <p className="text-sm text-muted-foreground mt-2">Use your Google account to continue</p>
          </div>

          <div className="flex min-h-10 justify-center" ref={googleButtonRef} />
          {isSigningIn && <p className="mt-4 text-center text-sm text-muted-foreground">Signing you in...</p>}
          {error && <p className="mt-4 text-center text-sm text-destructive">{error}</p>}
          <div className="mt-6 border-t border-border pt-5 text-center text-xs text-muted-foreground">
            Authentication is required to access the workspace.
          </div>
        </div>
      </div>
    </div>
  )
}
