"use client";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";

export default function StatusBar() {
  const { token } = useAuth();
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    setConnected(!!token);
  }, [token]);

  return (
    <div className="flex items-center gap-2 text-sm">
      <span className={`inline-block w-2.5 h-2.5 rounded-full ${connected ? "bg-emerald-500" : "bg-rose-500"}`} />
      <span className="text-slate-400">{connected ? "已连接" : "未连接"}</span>
    </div>
  );
}