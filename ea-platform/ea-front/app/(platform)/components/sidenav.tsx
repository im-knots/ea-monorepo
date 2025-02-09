import Link from 'next/link';
import Image from 'next/image';
import { Box } from '@mui/material';
import NavLinks from './navlinks';

export default function SideNav() {
  return (
    <Box className="flex h-full flex-col bg-gray-800 px-3 py-4 md:px-2">
      <Link
        className="mb-2 flex h-20 items-center justify-center md:h-40"
        href="/"
      >
        <Box className="w-32 text-white md:w-40">
          <Image src={'/logo.png'} alt="eru-logo" width={500} height={500} priority={true} />
        </Box>
      </Link>
      <Box className="flex grow flex-row justify-between space-x-2 md:flex-col md:space-x-0 md:space-y-2">
        <NavLinks/>
      </Box>
    </Box>
  );
}