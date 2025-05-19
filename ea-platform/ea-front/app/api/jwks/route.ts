import { NextResponse } from 'next/server';
import { getJWKSFile } from '@/app/lib/jwks';

export async function GET(req: Request) {
  const jwks = await getJWKSFile();
  return NextResponse.json(jwks);
}