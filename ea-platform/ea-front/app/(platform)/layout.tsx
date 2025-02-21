"use client";

import { useState } from "react";
import SideNav from "@/app/(platform)/components/sidenav";
import { Box } from "@mui/material";

export default function Layout({ children }: { children: React.ReactNode }) {
  const [isCollapsed, setIsCollapsed] = useState(false); // Manage the sidebar collapse state

  // Function to toggle the sidebar's collapsed state
  const toggleSidebar = () => setIsCollapsed((prev) => !prev);

  return (
    <Box className="flex h-screen flex-col md:flex-row md:overflow-hidden">
      <Box className="w-full flex-none md:w-64">
        {/* Pass isCollapsed and toggleSidebar to the SideNav component */}
        <SideNav isCollapsed={isCollapsed} toggleSidebar={toggleSidebar} />
      </Box>
      <Box className="flex-grow md:overflow-y-auto">{children}</Box>
    </Box>
  );
}
