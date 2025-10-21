import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:3281/api/:path*",
      },
      {
        source: "/files/:path*",
        destination: "http://localhost:3281/files/:path*",
      },
      {
        source: "/upload/:path*",
        destination: "http://localhost:3281/upload/:path*",
      },
      {
        source: "/tus/:path*",
        destination: "http://localhost:3281/tus/:path*",
      },
      {
        source: "/ws",
        destination: "http://localhost:3281/ws",
      },
    ];
  },
  experimental: {
    allowedDevOrigins: ["192.168.1.7:3000"]
  }
};

export default nextConfig;
