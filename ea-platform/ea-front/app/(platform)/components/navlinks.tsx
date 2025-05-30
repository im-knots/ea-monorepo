'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import HomeIcon from '@mui/icons-material/Home';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import GroupAddIcon from '@mui/icons-material/GroupAdd';
import SettingsIcon from '@mui/icons-material/Settings';
import SmartToyIcon from '@mui/icons-material/SmartToy';
import LocalGroceryStoreIcon from '@mui/icons-material/LocalGroceryStore';
import LeaderboardIcon from '@mui/icons-material/Leaderboard';
import HelpCenterIcon from '@mui/icons-material/HelpCenter';
import LogoutIcon from '@mui/icons-material/Logout';
import DeveloperBoardIcon from '@mui/icons-material/DeveloperBoard';
import clsx from 'clsx';
import { usePathname } from 'next/navigation';
import { Box } from '@mui/material';

const links = [
  { name: 'Dashboard', href: '/dashboard', icon: HomeIcon },
  { name: 'Agent Builder', href: '/agent-builder', icon: SmartToyIcon },
  { name: 'Agent Manager', href: '/agent-manager', icon: GroupAddIcon },
  { name: 'Node Builder', href: '/node-builder', icon: DeveloperBoardIcon},
  { name: 'Datasets', href: '/datasets', icon: ContentCopyIcon},
  { name: 'Marketplace', href: '/marketplace', icon: LocalGroceryStoreIcon },
  { name: 'Leaderboards', href: '/leaderboards', icon: LeaderboardIcon },
  { name: 'Settings', href: '/settings', icon: SettingsIcon },
  { name: 'Help', href: '/help', icon: HelpCenterIcon },
];

export default function NavLinks() {
  const pathname = usePathname();
  const router = useRouter();

  // Handle Logout Function
  const handleLogout = async () => {
    try {
      const response = await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include', // Ensures cookies are sent
      });
  
      const data = await response.json();
      console.log("🔹 Logout API Response:", data);
  
      if (response.ok) {
        router.push('/login'); // Redirect after successful logout
      } else {
        console.error('❌ Logout failed:', data);
      }
    } catch (error) {
      console.error('❌ Logout request error:', error);
    }
  };
  

  return (
    <Box>
      {links.map((link) => {
        const LinkIcon = link.icon;
        return (
          <Link
            key={link.name}
            href={link.href}
            className={clsx(
              "flex h-[48px] grow items-center justify-center rounded-[0.5vw] gap-2 p-3 text-sm font-medium hover:bg-gray-700 md:flex-none md:justify-start md:p-2 md:px-3",
              {
                'bg-gray-700': link.href === pathname,
              }
            )}
          >
            <LinkIcon className="w-6" />
            <p className="hidden md:block">{link.name}</p>
          </Link>
        );
      })}

      {/* Logout Button (Not a Link) */}
      <button
        onClick={handleLogout}
        className="flex w-full h-[48px] items-center justify-center rounded-[0.5vw] gap-2 p-3 text-sm font-medium hover:bg-gray-700 md:flex-none md:justify-start md:p-2 md:px-3"
      >
        <LogoutIcon className="w-6" />
        <p className="hidden md:block">Sign Out</p>
      </button>
    </Box>
  );
}
