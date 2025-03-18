"use client";

import { Box, Button, TextField, Typography } from "@mui/material";
import Image from "next/image";
import { register } from "../lib/actions";
import { useActionState } from "react";
import { redirect } from "next/navigation";

export default function Page() {

  const [ errorMessage, formAction, isPending ] = useActionState(
    register,
    undefined
  );

  return (
      <Box id="root" className="bg-neutral-900 text-white h-screen flex flex-col justify-center items-center">
        <Box className="flex flex-col justify-center items-center p-6 bg-neutral-800 rounded-lg shadow-lg w-[350px]">
          <Box className="w-32 md:w-40">
            <Image src={'/logo.png'} alt="eru-logo" width={500} height={500} priority={true} />
          </Box>
          <Typography variant="h6">Register</Typography>
          <form action={formAction} className="p-4 w-full flex flex-col justify-center items-center gap-4">
            <TextField
              id="email"
              name="email"
              label="Email"
              fullWidth
              className="bg-neutral-700 rounded-md"
            />
            <TextField
              id="password"
              name="password"
              label="Password"
              type="password"
              fullWidth
              className="bg-neutral-700 rounded-md"
            />
            <TextField
              id="alphaCode"
              name="alphaCode"
              label="Alpha Code"
              type="password"
              fullWidth
              className="bg-neutral-700 rounded-md"
            />
            <Button type="submit" variant="contained" color="primary" className="w-full">Register</Button>
            <Button variant='contained' color='secondary' className='w-full' onClick={() => redirect('/login')}>Login</Button>
          </form>
          {errorMessage && (
            <>
              <Typography variant="body1" className="text-sm text-red-500">{errorMessage}</Typography>
            </>
          )}
        </Box>
      </Box>
  )
}