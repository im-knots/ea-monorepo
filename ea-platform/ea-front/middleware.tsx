import { NextRequest, NextResponse } from 'next/server';
import { jwtVerify } from 'jose';

export async function middleware(req: NextRequest) {
  if (!process.env.JWT_SECRET) {
    throw new Error('JWT_SECRET not set');
  }
  const token = req.cookies.get('token')?.value;
  const JWT_SECRET_KEY = new TextEncoder().encode(process.env.JWT_SECRET);

  const unprotectedRoutes = ['/login', '/register'];
  const isProtectedRoute = !unprotectedRoutes.includes(req.nextUrl.pathname);

  try {
    if (!token && isProtectedRoute) {
      console.log('No valid token. Redirecting to /login...');
      return NextResponse.redirect(new URL('/login', req.url));
    }
    if (token) {
      jwtVerify(token, JWT_SECRET_KEY);
      return NextResponse.next();
    }
  } catch (error) {
    console.log('Invalid or expired token:', error);
    return NextResponse.redirect(new URL('/login', req.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: [
     '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ]
}