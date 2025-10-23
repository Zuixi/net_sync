"use client";
import { createContext, useContext, useEffect, useState, useCallback, useRef, ReactNode } from "react";
import { useAuth } from "./auth";

export interface ChatMessage {
  type: string;
  id?: string;
  from?: string;
  text?: string;
  timestamp?: number;
  device?: string;
}

interface WebSocketContextType {
  connected: boolean;
  messages: ChatMessage[];
  sendMessage: (message: any) => void;
  clearMessages: () => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within WebSocketProvider");
  }
  return context;
}

interface WebSocketProviderProps {
  children: ReactNode;
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const { token } = useAuth();
  const [connected, setConnected] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const [deviceName] = useState(
    navigator.userAgent.includes("Mobile") ? "移动设备" : "网页浏览器"
  );

  const sendMessage = useCallback((message: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket not connected, cannot send message");
    }
  }, []);

  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  useEffect(() => {
    console.log("WebSocket useEffect triggered, token:", token ? "存在" : "null");
    if (!token) {
      console.log("没有token,跳过WebSocket连接");
      setConnected(false);
      return;
    }

    console.log("开始建立WebSocket连接...");
    const connectWebSocket = () => {
      // 尝试通过相对路径连接(在同一端口),如果失败则直接连接到3280
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const host = window.location.hostname;
      const port = "3280"; // 直接连接到Go服务器
      const wsUrl = `${protocol}//${host}:${port}/ws?token=${encodeURIComponent(token)}`;

      console.log("WebSocket URL:", wsUrl);
      const socket = new WebSocket(wsUrl);
      wsRef.current = socket;

      socket.onopen = () => {
        console.log("WebSocket connected");
        setConnected(true);
        setMessages((m) => [{ type: "system", text: "已连接" }, ...m]);

        // 发送hello消息
        socket.send(JSON.stringify({
          type: "hello",
          device: deviceName,
          capabilities: ["upload", "download", "chat"],
        }));
      };

      socket.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data) as ChatMessage;
          setMessages((m) => [msg, ...m].slice(0, 200));
        } catch {
          // 如果不是JSON,当作文本消息处理
          setMessages((m) => [{ type: "text", text: ev.data }, ...m].slice(0, 200));
        }
      };

      socket.onerror = (error) => {
        console.error("WebSocket error:", error);
        setConnected(false);
      };

      socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
        setConnected(false);
        setMessages((m) => [{ type: "system", text: "连接断开" }, ...m]);
        wsRef.current = null;

        // 自动重连(除非是正常关闭)
        if (event.code !== 1000) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log("Attempting to reconnect...");
            connectWebSocket();
          }, 3000);
        }
      };
    };

    connectWebSocket();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close(1000, "Component unmounted");
        wsRef.current = null;
      }
    };
  }, [token, deviceName]);

  return (
    <WebSocketContext.Provider value={{ connected, messages, sendMessage, clearMessages }}>
      {children}
    </WebSocketContext.Provider>
  );
}
