import { NextResponse } from 'next/server';

export async function POST() {
  const response = NextResponse.json({ message: 'Logged out successfully' }, { status: 200 });

  // Remove the JWT token by setting it to an empty string and expiring it
  response.cookies.set('token', '', {
    httpOnly: true,
    expires: new Date(0), // Immediately expire the cookie
    path: '/',
  });

  return response;
}
