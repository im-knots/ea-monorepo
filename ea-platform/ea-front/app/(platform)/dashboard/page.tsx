import { Box, Typography } from "@mui/material";
import CardWrapper from "../components/cards";

export default function Page() {
  return (
    <Box id="root" className="flex flex-col p-6">
      <Box id="header" className="mb-8">
        <Typography variant="h4">Dashboard</Typography>
      </Box>
      <Box id="cards">
        <CardWrapper />
      </Box>
      <Box id="compute"></Box>
      <Box id="agents"></Box>
      <Box id="jobs"></Box>
    </Box>
  )
}