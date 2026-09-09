import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';
const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;

export async function POST(request) {
  try {
    const body = await request.json();
    const { credential } = body;

    if (!credential || typeof credential !== 'string') {
      return NextResponse.json(
        { error: 'Google credential is required' },
        { status: 400 }
      );
    }

    const credentialParts = credential.split('.');
    if (credentialParts.length !== 3) {
      return NextResponse.json(
        { error: 'Invalid Google credential' },
        { status: 401 }
      );
    }

    if (!GOOGLE_CLIENT_ID) {
      return NextResponse.json(
        { error: 'Google sign-in is not configured' },
        { status: 500 }
      );
    }

    const googleRes = await fetch(
      `https://oauth2.googleapis.com/tokeninfo?id_token=${encodeURIComponent(credential)}`,
      { cache: 'no-store' }
    );

    if (!googleRes.ok) {
      return NextResponse.json(
        { error: 'Google credential could not be verified' },
        { status: 401 }
      );
    }

    let profile;
    try {
      profile = await googleRes.json();
    } catch {
      return NextResponse.json(
        { error: 'Invalid Google credential' },
        { status: 401 }
      );
    }

    if (profile.aud !== GOOGLE_CLIENT_ID) {
      return NextResponse.json(
        { error: 'Google credential belongs to a different client' },
        { status: 401 }
      );
    }

    const { name, email, picture, sub: googleId } = profile;
    if (!name || !email || !googleId) {
      return NextResponse.json(
        { error: 'Google profile is incomplete' },
        { status: 401 }
      );
    }

    const backendRes = await fetch(`${BACKEND_URL}/user`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name, email, picture, googleId }),
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
      tokenString = data.tokenString || data.token;
    } else {
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
      maxAge: 60 * 60 * 24 * 7,
    });
    cookieStore.set({
      name: 'userProfile',
      value: Buffer.from(JSON.stringify({ name, email, picture })).toString('base64url'),
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      path: '/',
      maxAge: 60 * 60 * 24 * 7,
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
