import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Easy Sync",
  description: "Transfer files between devices on your local network",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" className="h-full">
      <head>
        <link rel="manifest" href="/manifest.json" />
      </head>
      <body className={`${inter.className} h-full bg-slate-900 text-slate-100`}>{children}</body>
    </html>
  );
}
