"use client";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";
import { useAutoPair } from "@/lib/auto-pair";

export default function StatusBar() {
  const { token } = useAuth();
  const { loading: pairLoading, autoPaired } = useAutoPair();
  const [wsConnected, setWsConnected] = useState(false);

  // 检查WebSocket连接状态
  useEffect(() => {
    if (!token) {
      setWsConnected(false);
      return;
    }

    // 尝试连接WebSocket来验证token是否有效
    const url = new URL(window.location.origin);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    url.pathname = "/ws";
    url.searchParams.set("token", token);

    const ws = new WebSocket(url.toString());

    ws.onopen = () => {
      setWsConnected(true);
    };

    ws.onerror = () => {
      setWsConnected(false);
    };

    ws.onclose = () => {
      setWsConnected(false);
    };

    return () => {
      ws.close();
    };
  }, [token]);

  const getStatus = () => {
    if (pairLoading) return { color: "bg-yellow-500", text: "连接中..." };
    if (wsConnected) return { color: "bg-emerald-500", text: "已连接" };
    if (token && !wsConnected) return { color: "bg-orange-500", text: "连接断开" };
    return { color: "bg-rose-500", text: "未连接" };
  };

  const status = getStatus();

  return (
    <div className="flex items-center gap-2 text-sm">
      <span className={`inline-block w-2.5 h-2.5 rounded-full ${status.color}`} />
      <span className="text-slate-400">{status.text}</span>
    </div>
  );
}