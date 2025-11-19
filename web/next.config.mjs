/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable standalone output for Docker builds
  output: 'standalone',

  // Rewrites for local development (not used in Docker where nginx handles the proxy)
  async rewrites() {
    // In local dev, proxy /api/* to the Go backend
    // Default to localhost:8080, can be overridden with API_URL env var
    const backendUrl = process.env.API_URL || 'http://localhost:8080';

    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
