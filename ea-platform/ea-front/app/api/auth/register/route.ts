import { NextResponse } from "next/server";
import { Register } from "@/app/lib/auth";


export async function POST(req: Request) {
  try {
    const { email, password, alphaCode } = await req.json();
    await Register({ email, password, alphaCode});
    return NextResponse.json({ message: "User registered" }, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: "Registration error" }, { status: 500 });
  }
}