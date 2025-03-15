import bcrypt from "bcryptjs";
import mongodb from "./mongodb";
import * as jose from "jose";
import { cookies } from "next/headers";
import { JWKS, privateKey } from "./jwks";

export async function Register(user: { email: string, password: string }) {
  const { email, password } = user;
  const hashedPassword = await bcrypt.hash(password, 10);
  const userRecord = await mongodb.db().collection("ainuUsers").findOne({ email });
  
  if (userRecord) {
    console.log("User already exists");
    throw new Error("User already exists");
  }
  await mongodb.db().collection("ainuUsers").insertOne({
    email,
    password: hashedPassword,
  });
}

export async function Login(user: { email: string, password: string }) {
  const cookieStore = await cookies();

  const { email, password } = user;
  
  if (!process.env.JWT_SECRET) {
    throw new Error("JWT_SECRET not set");
  }
  
  const userRecord = await mongodb.db().collection("ainuUsers").findOne({ email });

  if (!userRecord) {
    console.log("User not found");
    throw new Error("User not found");
  }

  const passwordMatch = await bcrypt.compare(password, userRecord.password);

  if (!passwordMatch) {
    console.log("Invalid password");
    throw new Error("Invalid password");
  }
  try {
    const token = await new jose.SignJWT({
      email: userRecord.email,
    })
    .setProtectedHeader({ alg: "RS256" })
    .setIssuedAt()
    .setExpirationTime("6h")
    .sign(privateKey);
  
    cookieStore.set('token', token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
    })
    return token;
  } catch (error) {
    console.log("Error creating token", error);
    throw new Error("Error creating token");
  }
}

export async function GetTokenClaims(token: string) {
  if (!token) {
    return {};
  }
  try {
    const { payload } = await jose.jwtVerify(token, JWKS);
    return payload;
  } catch (error) {
    console.log("Invalid token", error);
    return {};
  }
}