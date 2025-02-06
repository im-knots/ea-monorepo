import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from "@mui/material/Typography";
import TextField from '@mui/material/TextField';

export default function Page() {
  return (
    <Box className="" id="login-page">
      <Box className="absolute p-8" id="tagline">
        <Typography variant="h5">EruLabs</Typography>
      </Box>
      <Box className="h-screen flex flex-col justify-center items-center" id="login-container">
        <Box className="flex flex-col justify-center items-center" id="login-form">
          <Box className="p-4">
            <Typography variant="h6">Login</Typography>
          </Box>
          <Box className="p-4">
            <TextField 
                id="username"
                label="Username"
            />
          </Box>
          <Box className="p-4">
            <TextField
                id="password"
                label="Password"
                type="password"
                autoComplete="current-password"
            />
          </Box>
          <Box className="p-4">
            <Button>Login</Button>
          </Box>
        </Box>
      </Box>
    </Box>
  )
}