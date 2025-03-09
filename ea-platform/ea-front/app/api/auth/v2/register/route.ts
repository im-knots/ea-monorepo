import { NextResponse } from "next/server";
import { prisma } from "@/app/lib/prisma";
import bcrypt from "bcryptjs";
import { Register } from "@/app/lib/auth";
import { Prisma } from "@prisma/client";

export async function POST(req: Request) {
  try {
    const { email, password } = await req.json();
    Register({ email, password });
    return NextResponse.json({ message: "User registered" }, { status: 201 });
  } catch (error) {
    if (error instanceof Prisma.PrismaClientKnownRequestError) {
      if (error.code === "P2002") {
        return NextResponse.json({ error: "User already exists" }, { status: 409 });
      }
    }
    return NextResponse.json({ error: "Server error" }, { status: 500 });
  }
}