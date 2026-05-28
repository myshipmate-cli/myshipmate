/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/install', destination: '/install.sh' },
      { source: '/install.ps1', destination: '/install.ps1' },
    ];
  },
};

module.exports = nextConfig;
