"use client";
import { useRef } from "react";
import { useAuth } from "@/lib/auth";
import { useWebSocket, ChatMessage } from "@/lib/websocket-context";

export default function Chat() {
  const { token } = useAuth();
  const { messages, sendMessage, connected, deviceName } = useWebSocket();
  const inputRef = useRef<HTMLInputElement | null>(null);

  function send() {
    const v = inputRef.current?.value?.trim();
    if (!v || !connected) return;

    const msg = {
      type: "chat",
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      text: v,
      timestamp: Date.now(),
      from: deviceName, // 设置from字段为当前设备名
    };

    sendMessage(msg);
    inputRef.current!.value = "";
  }

  function formatMessage(msg: ChatMessage, index: number) {
    // 只显示 system 和 chat 类型的消息，忽略 presence、hello 等
    if (msg.type === "system") {
      return (
        <div key={index} className="text-center text-xs text-slate-500 py-1">
          {msg.text}
        </div>
      );
    }

    if (msg.type === "chat") {
      // 正确判断是否为自己发送的消息：比较 msg.from 和当前设备名
      const isFromSelf = msg.from === deviceName;
      return (
        <div key={index} className={`flex ${isFromSelf ? "justify-end" : "justify-start"} mb-2`}>
          <div className={`max-w-[75%] rounded-lg px-4 py-2.5 ${
            isFromSelf
              ? "bg-sky-600 text-white rounded-br-sm"
              : "bg-slate-700 text-slate-100 rounded-bl-sm"
          }`}>
            {!isFromSelf && msg.from && (
              <div className="text-xs text-slate-300 mb-1 font-medium">{msg.from}</div>
            )}
            <div className="text-sm leading-relaxed break-words">{msg.text}</div>
            {msg.timestamp && (
              <div className={`text-[10px] mt-1 ${isFromSelf ? "text-sky-200" : "text-slate-400"}`}>
                {new Date(msg.timestamp * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </div>
            )}
          </div>
        </div>
      );
    }

    // 忽略其他类型消息（presence、hello 等）
    return null;
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