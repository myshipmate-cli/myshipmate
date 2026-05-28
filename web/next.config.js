/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/install', destination: '/install.sh' },
    ];
  },
};

module.exports = nextConfig;
