"use client";
import { useRef } from "react";
import { useAuth } from "@/lib/auth";
import { useWebSocket, ChatMessage } from "@/lib/websocket-context";

export default function Chat() {
  const { token } = useAuth();
  const { messages, sendMessage, connected } = useWebSocket();
  const inputRef = useRef<HTMLInputElement | null>(null);

  function send() {
    const v = inputRef.current?.value?.trim();
    if (!v || !connected) return;

    const msg = {
      type: "chat",
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      text: v,
      timestamp: Date.now(),
    };

    sendMessage(msg);
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