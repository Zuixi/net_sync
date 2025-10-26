import type { NextConfig } from "next";

// 从环境变量读取服务器 URL，默认为 localhost:3280
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:3280";

// 从环境变量读取允许的开发源
const DEV_ALLOWED_ORIGIN = process.env.NEXT_PUBLIC_DEV_ALLOWED_ORIGIN;

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${API_BASE_URL}/api/:path*`,
      },
      {
        source: "/files/:path*",
        destination: `${API_BASE_URL}/files/:path*`,
      },
      {
        source: "/upload/:path*",
        destination: `${API_BASE_URL}/upload/:path*`,
      },
      {
        source: "/tus/:path*",
        destination: `${API_BASE_URL}/tus/:path*`,
      },
      {
        source: "/ws",
        destination: `${API_BASE_URL}/ws`,
      },
    ];
  },
  experimental: {
    // 仅在提供了开发源时才配置
    ...(DEV_ALLOWED_ORIGIN && {
      allowedDevOrigins: [DEV_ALLOWED_ORIGIN]
    })
  }
};

export default nextConfig;
