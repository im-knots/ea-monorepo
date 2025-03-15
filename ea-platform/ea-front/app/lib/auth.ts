import bcrypt from "bcryptjs";
import mongoClient from "./mongodb";
import * as jose from "jose";
import { cookies } from "next/headers";
import { getJWKS, getPrivateKey } from "./jwks";

export async function Register(user: { email: string, password: string }) {
  const { email, password } = user;
  const hashedPassword = await bcrypt.hash(password, 10);
  const mongodb = await mongoClient();
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

  const mongodb = await mongoClient();
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
    const privateKey = await getPrivateKey();
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
    const JWKS = await getJWKS();
    const { payload } = await jose.jwtVerify(token, JWKS);
    return payload;
  } catch (error) {
    console.log("Invalid token", error);
    return {};
  }
}