import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

export async function POST(request) {
  try {
    const body = await request.json();

    // Get the JWT from the HttpOnly cookie
    const cookieStore = await cookies();
    const tokenCookie = cookieStore.get('tokenString');
    const token = tokenCookie?.value;

    if (!token) {
      return NextResponse.json(
        { error: 'Unauthorized: No token found' },
        { status: 401 }
      );
    }

    // Send the logs to the Go backend
    const backendRes = await fetch(`${BACKEND_URL}/logs`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(body),
    });

    if (!backendRes.ok) {
      const errorText = await backendRes.text();
      return NextResponse.json(
        { error: 'Failed to send logs to backend', details: errorText },
        { status: backendRes.status }
      );
    }
    
    // Parse the response if there is any
    const contentType = backendRes.headers.get('content-type');
    let data = null;
    if (contentType && contentType.includes('application/json')) {
      data = await backendRes.json();
    } else {
      const text = await backendRes.text();
      if (text) data = text;
    }

    return NextResponse.json({ success: true, data }, { status: backendRes.status });
  } catch (error) {
    console.error('Error in logs API:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}
