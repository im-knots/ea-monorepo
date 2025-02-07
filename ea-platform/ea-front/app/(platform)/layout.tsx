import SideNav from "@/app/(platform)/components/sidenav";
import { Box } from "@mui/material";
 
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <Box className="flex h-screen flex-col md:flex-row md:overflow-hidden">
      <Box className="w-full flex-none md:w-64">
        <SideNav />
      </Box>
      <Box className="flex-grow p-6 md:overflow-y-auto md:p-12">{children}</Box>
    </Box>
  );
}
