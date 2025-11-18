/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable standalone output for Docker builds
  output: 'standalone',

  async rewrites() {
    // Use environment variable for backend URL to support Docker
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`, // Proxy to Go backend
      },
    ];
  },
};

export default nextConfig;
