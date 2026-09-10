import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080'

export async function GET(request, { params }) {
  try {
    const { id } = await params
    const token = (await cookies()).get('tokenString')?.value
    if (!token) return NextResponse.json({ error: 'Unauthorized: No token found' }, { status: 401 })

    const backendRes = await fetch(`${BACKEND_URL}/logs/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    const body = await backendRes.json().catch(() => ({ error: 'Invalid backend response' }))
    return NextResponse.json(body, { status: backendRes.status })
  } catch (error) {
    console.error('Error in log detail API:', error)
    return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 })
  }
}
