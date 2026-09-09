import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';

export async function GET() {
  const cookieStore = await cookies();
  const token = cookieStore.get('tokenString')?.value;
  const profileCookie = cookieStore.get('userProfile')?.value;

  if (!token || !profileCookie) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    const profile = JSON.parse(Buffer.from(profileCookie, 'base64url').toString('utf8'));
    return NextResponse.json(profile);
  } catch {
    return NextResponse.json({ error: 'Invalid user profile' }, { status: 401 });
  }
}