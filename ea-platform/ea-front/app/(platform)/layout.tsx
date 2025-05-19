"use client";

import { useEffect, useState } from "react";
import SideNav from "@/app/(platform)/components/sidenav";
import { Box } from "@mui/material";
import { useRouter } from "next/navigation";

export default function Layout({ children }: { children: React.ReactNode }) {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const toggleSidebar = () => setIsCollapsed((prev) => !prev);
  
  const router = useRouter();
  useEffect(() => {
    fetch("/api/user").then(async (res) => {
      if (res.status === 401) {
        router.push("/login");
      }
    });
  }, [])

  return (
    <Box className="flex h-screen flex-col md:flex-row md:overflow-hidden">
      {/* Sidebar width dynamically changes */}
      <Box className={`transition-all duration-300 ${isCollapsed ? "w-16" : "w-64"}`}>
        <SideNav isCollapsed={isCollapsed} toggleSidebar={toggleSidebar} />
      </Box>

      {/* Main content shifts accordingly */}
      <Box className={`flex-grow md:overflow-y-auto transition-all duration-300`}>
        {children}
      </Box>
    </Box>
  );
}
