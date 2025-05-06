import { Login } from "@/app/lib/auth";
import { NextResponse } from "next/server";


export async function POST(req: Request) {
  try {
    const { email, password } = await req.json();
    const jwt = await Login({ email, password });
    return NextResponse.json({ message: "Logged in", token: jwt }, { status: 201 });
  } catch (error) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
}