"use client";
import { useEffect, useRef, useState } from "react";
import { useAuth } from "@/lib/auth";

export default function Chat() {
  const { token } = useAuth();
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [messages, setMessages] = useState<string[]>([]);
  const inputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (!token) return;
    const url = new URL(window.location.origin);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    url.pathname = "/ws";
    url.searchParams.set("token", token);

    const socket = new WebSocket(url.toString());
    setWs(socket);

    socket.onmessage = (ev) => {
      setMessages((m) => [ev.data, ...m].slice(0, 200));
    };
    socket.onopen = () => setMessages((m) => ["[connected]", ...m]);
    socket.onclose = () => setMessages((m) => ["[disconnected]", ...m]);

    return () => socket.close();
  }, [token]);

  function send() {
    const v = inputRef.current?.value?.trim();
    if (!v || !ws) return;
    ws.send(v);
    inputRef.current!.value = "";
  }

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
      {!token && <p className="text-sm text-slate-400">请先完成配对</p>}
      {token && (
        <>
          <div className="flex gap-2">
            <input ref={inputRef} className="flex-1 rounded-md bg-slate-800 px-3 py-2 text-sm outline-none focus:ring-2 ring-sky-600" placeholder="输入消息" />
            <button onClick={send} className="rounded-md bg-sky-600 px-3 py-2 text-sm hover:bg-sky-500">发送</button>
          </div>
          <div className="mt-3 h-64 overflow-auto rounded-md bg-slate-800 p-3 text-xs space-y-1">
            {messages.map((m, i) => (<div key={i} className="rounded bg-slate-700/70 px-2 py-1">{m}</div>))}
          </div>
        </>
      )}
    </div>
  );
}