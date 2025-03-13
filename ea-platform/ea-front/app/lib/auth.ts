import bcrypt from "bcryptjs";
import mongodb from "./mongodb";
import jwt from "jsonwebtoken";
import { cookies } from "next/headers";

export async function Register(user: { email: string, password: string }) {
  const { email, password } = user;
  const hashedPassword = await bcrypt.hash(password, 10);
  const userRecord = await mongodb.db().collection("users").findOne({ email });
  
  if (userRecord) {
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
    throw new Error("User not found");
  }

  const passwordMatch = await bcrypt.compare(password, userRecord.password);

  if (!passwordMatch) {
    throw new Error("Invalid password");
  }
 
  const token = jwt.sign({
      userId: userRecord.id,
      email,
    },
    process.env.JWT_SECRET,
    {
      expiresIn: '6h'
    }
  );
  cookieStore.set('token', token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
  })
  return token;
}