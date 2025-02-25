'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from "@mui/material/Typography";
import TextField from '@mui/material/TextField';
import Image from 'next/image';

export default function Page() {
  const router = useRouter();
  const [isRegistering, setIsRegistering] = useState(false);
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [alphaCode, setAlphaCode] = useState(''); // New input state for Alpha Access Code
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('');

  const handleAuth = async () => {
    setMessage('');
    setMessageType('');

    const endpoint = isRegistering ? '/api/auth/register' : '/api/auth/login';
    const body = isRegistering ? { name, email, password, alphaCode } : { email, password };

    try {
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        credentials: 'include',
      });

      const data = await res.json();

      if (res.ok) {
        setMessage(isRegistering ? 'User created successfully!' : 'Login successful!');
        setMessageType('success');

        if (!isRegistering) {
          router.push('/dashboard'); // Redirect after login
        }
      } else {
        setMessage(data.error);
        setMessageType('error');
      }
    } catch (error) {
      setMessage('Error processing request.');
      setMessageType('error');
    }
  };

  return (
    <Box className="bg-neutral-900 text-white h-screen flex flex-col justify-center items-center">
      <Box className="absolute p-8"></Box>
      <Box className="flex flex-col justify-center items-center p-6 bg-neutral-800 rounded-lg shadow-lg w-[350px]">
        <Box className="w-32 md:w-40">
          <Image src={'/logo.png'} alt="eru-logo" width={500} height={500} priority={true} />
        </Box>
        <Box className="p-4">
          <Typography variant="h6">{isRegistering ? 'Register' : 'Login'}</Typography>
        </Box>
        {isRegistering && (
          <>
            <Box className="p-4 w-full">
              <TextField 
                id="name" 
                label="Name" 
                value={name} 
                onChange={(e) => setName(e.target.value)} 
                fullWidth 
                className="bg-neutral-700 rounded-md"
              />
            </Box>
            <Box className="p-4 w-full">
              <TextField 
                id="alpha-code" 
                label="Alpha Access Code" 
                value={alphaCode} 
                onChange={(e) => setAlphaCode(e.target.value)} 
                fullWidth 
                className="bg-neutral-700 rounded-md"
              />
            </Box>
          </>
        )}
        <Box className="p-4 w-full">
          <TextField 
            id="email" 
            label="Email" 
            value={email} 
            onChange={(e) => setEmail(e.target.value)} 
            fullWidth 
            className="bg-neutral-700 rounded-md"
          />
        </Box>
        <Box className="p-4 w-full">
          <TextField 
            id="password" 
            label="Password" 
            type="password" 
            value={password} 
            onChange={(e) => setPassword(e.target.value)} 
            fullWidth 
            className="bg-neutral-700 rounded-md"
          />
        </Box>
        <Box className="p-4 w-full">
          <Button 
            onClick={handleAuth} 
            fullWidth 
            variant="contained" 
            sx={{ bgcolor: 'neutral.900', '&:hover': { bgcolor: 'neutral.700' } }}
          >
            {isRegistering ? 'Register' : 'Login'}
          </Button>
        </Box>
        {message && (
          <Typography 
            variant="body1" 
            sx={{ color: messageType === 'success' ? 'green' : 'red', mt: 2 }}
          >
            {message}
          </Typography>
        )}
        <Box className="p-4 w-full">
          <Button 
            onClick={() => setIsRegistering(!isRegistering)} 
            fullWidth 
            variant="outlined" 
            sx={{ borderColor: 'neutral.700', color: 'white', '&:hover': { bgcolor: 'neutral.700' } }}
          >
            {isRegistering ? 'Already have an account? Login' : 'Need an account? Register'}
          </Button>
        </Box>
      </Box>
    </Box>
  );
}
