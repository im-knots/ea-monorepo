import { NextResponse } from 'next/server';
import { GetTokenClaims } from '@/app/lib/auth';
import mongoClient from '@/app/lib/mongodb';
import { cookies } from 'next/headers';


export async function GET(req: Request) {
  const cookieStore = await cookies();
  const token = cookieStore.get('token')?.value;
  if (!token) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }
  const payload = await GetTokenClaims(token);
  const mongodb = await mongoClient();
  const user = await mongodb.db().collection("users").findOne({ email: payload.email });
  if (!user) {
    return NextResponse.json({ error: 'User not found' }, { status: 401 });
  }
  return NextResponse.json({
    email: user.email,
    id: user.id
  });
}