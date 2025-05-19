'use server';

import { z } from 'zod';
import { Login, Register } from './auth';
import { redirect } from 'next/navigation';

// Separate form schemas
const RegisterForm = z.object({
  email: z.string().email(),
  password: z.string(),
  alphaCode: z.string(),
});

const LoginForm = z.object({
  email: z.string().email(),
  password: z.string(),
});

export async function register(prevState: string | undefined, formData: FormData) {
  const validatedRegisterForm = RegisterForm.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
    alphaCode: formData.get('alphaCode')
  });

  if (!validatedRegisterForm.success) {
    return "Invalid form data";
  }

  try {
    await Register({
      email: validatedRegisterForm.data.email,
      password: validatedRegisterForm.data.password,
      alphaCode: validatedRegisterForm.data.alphaCode
    });
    await Login({
      email: validatedRegisterForm.data.email,
      password: validatedRegisterForm.data.password
    });
  } catch (error) {
    console.error("Registration error:", error);
    return "Registration error";
  }
  redirect('/login');
}

export async function login(prevState: string | undefined, formData: FormData) {
  const validatedForm = LoginForm.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
  });
  if (!validatedForm.success) {
    return "Invalid form data";
  }
  try {
    await Login({
      email: validatedForm.data.email,
      password: validatedForm.data.password
    });  
  } catch (error) {
    console.error("Login error:", error);
    return "Login error";
  }
  redirect('/dashboard');
}
