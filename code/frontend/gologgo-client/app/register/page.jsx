'use client'

import Link from 'next/link'
import { Mail, Lock, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ThemeToggle } from '@/components/theme-toggle'

export default function RegisterPage() {
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
            <h1 className="text-2xl font-semibold tracking-tight">Create an account</h1>
            <p className="text-sm text-muted-foreground mt-2">Enter your details to get started with GoLogGo</p>
          </div>

          <form className="flex flex-col gap-4" onSubmit={(e) => { e.preventDefault(); window.location.href = '/' }}>
            <div className="space-y-2">
              <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="email">Email</label>
              <div className="relative">
                <Mail className="absolute left-3 top-2.5 size-4 text-muted-foreground" />
                <Input id="email" type="email" placeholder="name@example.com" className="pl-9" required />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="username">Username</label>
              <div className="relative">
                <User className="absolute left-3 top-2.5 size-4 text-muted-foreground" />
                <Input id="username" placeholder="johndoe" className="pl-9" required />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="password">Password</label>
              <div className="relative">
                <Lock className="absolute left-3 top-2.5 size-4 text-muted-foreground" />
                <Input id="password" type="password" placeholder="••••••••" className="pl-9" required />
              </div>
            </div>

            <Button type="submit" className="w-full mt-2">Sign up</Button>
          </form>

          <div className="mt-6 text-center text-sm text-muted-foreground">
            Already have an account?{' '}
            <Link href="/login" className="font-medium text-foreground underline-offset-4 hover:underline">
              Sign in
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
