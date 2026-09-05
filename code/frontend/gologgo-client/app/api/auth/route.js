import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

export async function POST(request) {
  try {
    const body = await request.json();
    const { name, email } = body;

    if (!name || !email) {
      return NextResponse.json(
        { error: 'Name and email are required' },
        { status: 400 }
      );
    }

    // Send the user data to the Go backend
    const backendRes = await fetch(`${BACKEND_URL}/user`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, email }),
    });

    if (!backendRes.ok) {
      const errorText = await backendRes.text();
      return NextResponse.json(
        { error: 'Failed to authenticate with backend', details: errorText },
        { status: backendRes.status }
      );
    }

    // Parse the response from Go backend
    let tokenString;
    const contentType = backendRes.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      const data = await backendRes.json();
      // Assuming the backend might return { tokenString: "..." } or { token: "..." }
      tokenString = data.tokenString || data.token;
    } else {
      // If it returns plain text
      tokenString = await backendRes.text();
    }

    if (!tokenString) {
      return NextResponse.json(
        { error: 'No token received from backend' },
        { status: 500 }
      );
    }

    // Save the token as an HttpOnly cookie
    const cookieStore = await cookies();
    cookieStore.set({
      name: 'tokenString',
      value: tokenString.trim(),
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      path: '/',
      maxAge: 60 * 60 * 24 * 7, // 7 days (optional, remove if you want a session cookie)
    });

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error('Error in auth API:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}
