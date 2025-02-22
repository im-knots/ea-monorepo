"use client";

import { useState } from "react";
import SideNav from "@/app/(platform)/components/sidenav";
import { Box } from "@mui/material";

export default function Layout({ children }: { children: React.ReactNode }) {
  const [isCollapsed, setIsCollapsed] = useState(false);

  const toggleSidebar = () => setIsCollapsed((prev) => !prev);

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
