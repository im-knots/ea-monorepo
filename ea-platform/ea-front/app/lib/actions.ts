'use server';

import { z } from 'zod';
import { Login, Register } from './auth';
import { redirect } from 'next/navigation';

const LoginForm = z.object({
  email: z.string().email(),
  password: z.string(),
});

export async function register(prevState: string | undefined, formData: FormData) {
  const validatedForm = LoginForm.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
  });
  if (!validatedForm.success) {
    return "Invalid form data";
  }
  try {
    await Register({
      email: validatedForm.data.email,
      password: validatedForm.data.password
    });
    await Login({
      email: validatedForm.data.email,
      password: validatedForm.data.password
    })
  } catch (error) {
     return "Registration error";
  }
  redirect('/dashboard');
}

export async function login(prevState: string | undefined,formData: FormData) {
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
     return "Login error";
  }
  redirect('/dashboard');
}