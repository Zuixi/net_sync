"use client";
import { useAuth } from "@/lib/auth";
import { useAutoPair } from "@/lib/auto-pair";
import { useWebSocket } from "@/lib/websocket-context";

export default function StatusBar() {
  const { token } = useAuth();
  const { loading: pairLoading } = useAutoPair();
  const { connected } = useWebSocket();

  const getStatus = () => {
    if (pairLoading) return { color: "bg-yellow-500", text: "连接中..." };
    if (connected) return { color: "bg-emerald-500", text: "已连接" };
    if (token && !connected) return { color: "bg-orange-500", text: "连接断开" };
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