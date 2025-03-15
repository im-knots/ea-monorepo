import { NextResponse } from 'next/server';
import { jwksFile } from '@/app/lib/jwks';

export async function GET(req: Request) {
  return NextResponse.json(JSON.parse(jwksFile));
}