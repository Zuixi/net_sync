"use client";
import { useEffect, useRef, useState } from "react";
import { useAuth } from "@/lib/auth";

interface ChatMessage {
  type: string;
  id?: string;
  from?: string;
  text?: string;
  timestamp?: number;
  device?: string;
}

export default function Chat() {
  const { token } = useAuth();
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [deviceName] = useState(
    navigator.userAgent.includes("Mobile") ? "移动设备" : "网页浏览器"
  );
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
      try {
        const msg = JSON.parse(ev.data) as ChatMessage;
        setMessages((m) => [msg, ...m].slice(0, 200));
      } catch {
        // 如果不是JSON,当作文本消息处理
        setMessages((m) => [{ type: "text", text: ev.data }, ...m].slice(0, 200));
      }
    };
    socket.onopen = () => {
      setMessages((m) => [{ type: "system", text: "已连接" }, ...m]);
      // 发送hello消息
      socket.send(JSON.stringify({
        type: "hello",
        device: deviceName,
        capabilities: ["upload", "download", "chat"],
      }));
    };
    socket.onclose = () => setMessages((m) => [{ type: "system", text: "连接断开" }, ...m]);

    return () => socket.close();
  }, [token, deviceName]);

  function send() {
    const v = inputRef.current?.value?.trim();
    if (!v || !ws) return;

    const msg = {
      type: "chat",
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      text: v,
      timestamp: Date.now(),
    };

    ws.send(JSON.stringify(msg));
    inputRef.current!.value = "";
  }

  function formatMessage(msg: ChatMessage, index: number) {
    if (msg.type === "system") {
      return (
        <div key={index} className="text-center text-xs text-slate-500 py-1">
          {msg.text}
        </div>
      );
    }

    if (msg.type === "chat") {
      const isFromSelf = !msg.from;
      return (
        <div key={index} className={`flex ${isFromSelf ? "justify-end" : "justify-start"}`}>
          <div className={`max-w-[80%] rounded-lg px-3 py-2 ${
            isFromSelf
              ? "bg-sky-600/30 text-sky-100"
              : "bg-slate-700/70 text-slate-100"
          }`}>
            {msg.from && (
              <div className="text-xs text-slate-400 mb-1">{msg.from}</div>
            )}
            <div className="text-sm">{msg.text}</div>
          </div>
        </div>
      );
    }

    // 其他类型消息
    return (
      <div key={index} className="rounded bg-slate-700/70 px-2 py-1 text-xs text-slate-300">
        {msg.text || JSON.stringify(msg)}
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
      {!token && <p className="text-sm text-slate-400">请先完成配对</p>}
      {token && (
        <>
          <div className="flex gap-2 mb-3">
            <input
              ref={inputRef}
              className="flex-1 rounded-md bg-slate-800 px-3 py-2 text-sm outline-none focus:ring-2 ring-sky-600"
              placeholder="输入消息"
              onKeyDown={(e) => e.key === "Enter" && send()}
            />
            <button onClick={send} className="rounded-md bg-sky-600 px-3 py-2 text-sm hover:bg-sky-500">
              发送
            </button>
          </div>
          <div className="h-96 overflow-auto rounded-md bg-slate-800 p-3 space-y-2 flex flex-col-reverse">
            {messages.map((m, i) => formatMessage(m, i))}
          </div>
        </>
      )}
    </div>
  );
}