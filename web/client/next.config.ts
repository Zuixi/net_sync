import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:3280/api/:path*",
      },
      {
        source: "/files/:path*",
        destination: "http://localhost:3280/files/:path*",
      },
      {
        source: "/upload/:path*",
        destination: "http://localhost:3280/upload/:path*",
      },
      {
        source: "/tus/:path*",
        destination: "http://localhost:3280/tus/:path*",
      },
    ];
  },
};

export default nextConfig;
