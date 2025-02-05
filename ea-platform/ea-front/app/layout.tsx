import { AppRouterCacheProvider } from '@mui/material-nextjs/v15-appRouter';
import '@/app/global.css';
import { Roboto } from 'next/font/google';
import { ThemeProvider } from '@mui/material/styles';
import theme from './theme';
import { CssBaseline } from '@mui/material';

const roboto = Roboto({
  weight: ['300', '400', '500', '700'],
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-roboto',
});

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={roboto.variable}>
         <AppRouterCacheProvider options={{ key: 'css' }}>
            <ThemeProvider theme={theme}>
              <CssBaseline />
              {children}
            </ThemeProvider>
         </AppRouterCacheProvider>
      </body>
    </html>
  );
}