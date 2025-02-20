"use client"; // Add this line at the top

import { useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Box } from '@mui/material';
import NavLinks from './navlinks';
import { ChevronLeft, ChevronRight } from 'lucide-react';

export default function SideNav({ isCollapsed, toggleSidebar }: { isCollapsed: boolean; toggleSidebar: () => void }) {
  return (
    <Box className={`flex h-full flex-col bg-neutral-800 text-white px-3 py-4 ${isCollapsed ? 'w-16' : 'w-64'} transition-all`}>
      <button
        onClick={toggleSidebar}
        className="absolute top-4 left-4 p-2 bg-neutral-700 hover:bg-neutral-600 transition z-10"
      >
        {isCollapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
      </button>

      <Link className="mb-2 flex items-center justify-center" href="/">
        <Box className={`w-32 text-white ${isCollapsed ? 'hidden' : 'block'}`}>
          <Image src={'/logo.png'} alt="eru-logo" width={500} height={500} priority={true} />
        </Box>
      </Link>

      <Box className={`flex grow flex-row justify-between space-x-2 md:flex-col md:space-x-0 md:space-y-2 ${isCollapsed ? 'hidden' : 'block'}`}>
        <NavLinks />
      </Box>
    </Box>
  );
}
