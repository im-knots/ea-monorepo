import { NextResponse } from "next/server";
import { Register } from "@/app/lib/auth";

export async function POST(req: Request) {
  try {
    const { email, password } = await req.json();
    await Register({ email, password });
    return NextResponse.json({ message: "User registered" }, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: "Registration error" }, { status: 500 });
  }
}