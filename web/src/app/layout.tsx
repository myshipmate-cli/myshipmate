import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Shipmate - The Smart Deployer for Developers',
  description: 'Deploy any project to any platform with one command. Shipmate detects your project type and deploys it to the right platform automatically.',
  keywords: ['deploy', 'vercel', 'railway', 'render', 'netlify', 'fly.io', 'cli', 'developer tools'],
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
