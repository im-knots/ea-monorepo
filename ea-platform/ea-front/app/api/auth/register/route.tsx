import { NextResponse } from 'next/server';
import bcrypt from 'bcryptjs';
import { v4 as uuidv4 } from 'uuid';
import { connectToDatabase } from '@/lib/mongodb';

const ALPHA_ACCESS_CODE = 'some-alpha-string';

export async function POST(req: Request) {
  try {
    const { name, email, password, alphaCode } = await req.json();

    if (!name || !email || !password || !alphaCode) {
      return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
    }

    if (alphaCode !== ALPHA_ACCESS_CODE) {
      return NextResponse.json({ error: 'Invalid Alpha Access Code' }, { status: 403 });
    }

    const { db } = await connectToDatabase();
    const existingUser = await db.collection('users').findOne({ email });

    if (existingUser) {
      return NextResponse.json({ error: 'User already exists' }, { status: 400 });
    }

    const hashedPassword = await bcrypt.hash(password, 10);
    const userId = uuidv4();

    const newUser = {
      id: userId,
      name,
      email,
      password: hashedPassword,
      compute_credits: 0,
      compute_devices: [],
      jobs: [],
      created_time: new Date(),
    };

    await db.collection('users').insertOne(newUser);

    return NextResponse.json({ 
      message: 'User created successfully', 
      id: userId, 
      name: newUser.name, 
      created_time: newUser.created_time 
    }, { status: 201 });

  } catch (error) {
    console.error('Error in /register:', error);
    return NextResponse.json({ error: 'Server error' }, { status: 500 });
  }
}
