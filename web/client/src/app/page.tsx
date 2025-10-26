"use client";
import { useMemo, useState } from "react";
import { ArrowUpTrayIcon, ArrowDownTrayIcon, ChatBubbleLeftRightIcon, DevicePhoneMobileIcon } from "@heroicons/react/24/outline";
import Pairing from "@/components/Pairing";
import StatusBar from "@/components/StatusBar";
import Uploader from "@/components/Uploader";
import FileList from "@/components/FileList";
import Chat from "@/components/Chat";
import Devices from "@/components/Devices";
import { AuthProvider, useAuth } from "@/lib/auth";
import { useAutoPair } from "@/lib/auto-pair";
import { WebSocketProvider } from "@/lib/websocket-context";

function HomeContent() {
  const { token, loading: authLoading } = useAuth();
  const { loading: pairLoading, error: pairError } = useAutoPair();
  const [tab, setTab] = useState<"upload" | "download" | "chat" | "devices">("upload");
  const [showManualPairing, setShowManualPairing] = useState(false);

  const tabs = useMemo(() => ([
    { key: "upload" as const, label: "上传", icon: ArrowUpTrayIcon },
    { key: "download" as const, label: "下载", icon: ArrowDownTrayIcon },
    { key: "chat" as const, label: "聊天", icon: ChatBubbleLeftRightIcon },
    { key: "devices" as const, label: "设备", icon: DevicePhoneMobileIcon },
  ]), []);

  return (
    <div className="min-h-screen flex flex-col">
      <header className="sticky top-0 z-50 backdrop-blur bg-slate-900/70 border-b border-slate-800">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center gap-3">
          <h1 className="text-xl font-semibold">Easy Sync</h1>
          <div className="ml-auto"><StatusBar /></div>
        </div>
      <nav className="max-w-5xl mx-auto px-4 pb-3 flex gap-2">
        {tabs.map(({ key, label, icon: Icon }) => (
          <button key={key} onClick={() => setTab(key)}
            className={`inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm transition
            ${tab === key ? "bg-sky-600/20 text-sky-400 border border-sky-700" : "hover:bg-slate-800"}`}>
            <Icon className="w-4 h-4" /> {label}
          </button>
        ))}
      </nav>
    </header>

    <main className="max-w-5xl mx-auto w-full px-4 py-6 flex-1">
      {authLoading && <p className="text-slate-400">正在加载...</p>}
      {!authLoading && pairLoading && <p className="text-slate-400">正在自动连接...</p>}
      {!authLoading && pairError && !showManualPairing && (
        <div className="mb-6 rounded-lg border border-rose-800 bg-rose-900/20 p-4">
          <p className="text-sm text-rose-400 mb-2">自动配对失败: {pairError}</p>
          <button
            onClick={() => setShowManualPairing(true)}
            className="rounded-md bg-sky-600 px-3 py-2 text-sm hover:bg-sky-500"
          >
            手动配对
          </button>
        </div>
      )}
      {!authLoading && !token && showManualPairing && (
        <div className="mb-6">
          <Pairing />
        </div>
      )}

      {!authLoading && token && (
        <>
          {tab === "upload" && <Uploader />}
          {tab === "download" && <FileList />}
          {tab === "chat" && <Chat />}
          {tab === "devices" && <Devices />}
        </>
      )}
    </main>

      <footer className="sticky bottom-0 bg-slate-900/70 backdrop-blur border-t border-slate-800">
        <div className="max-w-5xl mx-auto px-4 py-3 text-sm text-slate-400">
          本地网络文件同步与聊天 · Tailwind · Next.js
        </div>
      </footer>
    </div>
  );
}

export default function Home() {
  return (
    <AuthProvider>
      <WebSocketProvider>
        <HomeContent />
      </WebSocketProvider>
    </AuthProvider>
  );
}
