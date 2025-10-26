"use client";
import { createContext, useContext, useEffect, useState, useCallback, useRef, ReactNode } from "react";
import { useAuth } from "./auth";
import { getWebSocketURL, WEBSOCKET_CONFIG } from "./config";

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
  deviceName: string;
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
  const messageIdsRef = useRef<Set<string>>(new Set()); // 用于消息去重

  // 从 localStorage 读取设备名称
  const [deviceName] = useState(() => {
    if (typeof window === 'undefined') return "网页浏览器";
    return localStorage.getItem("easy_sync_device_name") ||
      (navigator.userAgent.includes("Mobile") ? "移动设备" : "网页浏览器");
  });

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
      const wsUrl = `${getWebSocketURL()}?token=${encodeURIComponent(token)}`;

      console.log("WebSocket URL:", wsUrl);
      const socket = new WebSocket(wsUrl);
      wsRef.current = socket;

      socket.onopen = () => {
        console.log("WebSocket connected");
        setConnected(true);

        // 添加系统消息，带唯一 ID
        const connectMsg = {
          type: "system",
          id: `sys_connected_${Date.now()}`,
          text: "已连接"
        };
        setMessages((m) => [connectMsg, ...m]);

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

          // 消息去重：根据消息 ID 去重
          if (msg.id && messageIdsRef.current.has(msg.id)) {
            console.log("重复消息，忽略:", msg.id);
            return;
          }

          // 记录消息 ID
          if (msg.id) {
            messageIdsRef.current.add(msg.id);
            // 限制集合大小，防止内存泄漏
            if (messageIdsRef.current.size > WEBSOCKET_CONFIG.MESSAGE_HISTORY_LIMIT) {
              const firstId = messageIdsRef.current.values().next().value;
              messageIdsRef.current.delete(firstId);
            }
          }

          setMessages((m) => [msg, ...m].slice(0, WEBSOCKET_CONFIG.MESSAGE_HISTORY_LIMIT));
        } catch {
          // 如果不是JSON,当作文本消息处理
          setMessages((m) => [{ type: "text", text: ev.data }, ...m].slice(0, WEBSOCKET_CONFIG.MESSAGE_HISTORY_LIMIT));
        }
      };

      socket.onerror = (error) => {
        console.error("WebSocket error:", error);
        setConnected(false);
      };

      socket.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
        setConnected(false);

        // 添加系统消息，带唯一 ID
        const disconnectMsg = {
          type: "system",
          id: `sys_disconnected_${Date.now()}`,
          text: "连接断开"
        };
        setMessages((m) => [disconnectMsg, ...m]);
        wsRef.current = null;

        // 自动重连(除非是正常关闭)
        if (event.code !== 1000) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log("Attempting to reconnect...");
            connectWebSocket();
          }, WEBSOCKET_CONFIG.RECONNECT_DELAY);
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
    <WebSocketContext.Provider value={{ connected, messages, sendMessage, clearMessages, deviceName }}>
      {children}
    </WebSocketContext.Provider>
  );
}
