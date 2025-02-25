import { NextRequest, NextResponse } from 'next/server';
import { jwtVerify } from 'jose';

const JWT_SECRET = process.env.JWT_SECRET || 'supersecretkey';
const JWT_SECRET_KEY = new TextEncoder().encode(JWT_SECRET); // Convert to Uint8Array

export async function middleware(req: NextRequest) {
  const token = req.cookies.get('token')?.value;

  console.log('Middleware running...');
  console.log('Token:', token);
  console.log('Path:', req.nextUrl.pathname);

  const isAuthRoute = req.nextUrl.pathname === '/login';

  // Manually define a list of protected routes
  const protectedRoutes = [
    '/dashboard',
    '/agent-builder',
    '/agent-manager',
    '/node-builder',
    '/datasets',
  ];

  const isProtectedRoute = protectedRoutes.some((route) =>
    req.nextUrl.pathname.startsWith(route)
  );

  try {
    if (token) {
      const { payload } = await jwtVerify(token, JWT_SECRET_KEY);

      console.log('JWT Verified:', payload);

      // If logged-in user tries to access /login, redirect them to dashboard
      if (isAuthRoute) {
        console.log('User already logged in. Redirecting to /dashboard...');
        return NextResponse.redirect(new URL('/dashboard', req.url));
      }

      return NextResponse.next();
    }
  } catch (error) {
    console.log('Invalid or expired token:', error);
  }

  // If no valid token and trying to access a protected route, redirect to login
  if (!token && isProtectedRoute) {
    console.log('No valid token. Redirecting to /login...');
    return NextResponse.redirect(new URL('/login', req.url));
  }

  return NextResponse.next();
}
