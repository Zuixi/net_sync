"use client";
import { useEffect, useMemo, useState } from "react";
import { ArrowUpTrayIcon, ArrowDownTrayIcon, ChatBubbleLeftRightIcon, DevicePhoneMobileIcon } from "@heroicons/react/24/outline";
import Pairing from "@/components/Pairing";
import StatusBar from "@/components/StatusBar";
import Uploader from "@/components/Uploader";
import FileList from "@/components/FileList";
import Chat from "@/components/Chat";
import Devices from "@/components/Devices";
import { useApiConfig } from "@/lib/config";
import { useAuth } from "@/lib/auth";

export default function Home() {
  const { config, loading: cfgLoading, error: cfgError } = useApiConfig();
  const { token } = useAuth();
  const [tab, setTab] = useState<"upload" | "download" | "chat" | "devices">("upload");

  useEffect(() => {
    if (cfgError) console.error("Config load error", cfgError);
  }, [cfgError]);

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
        {cfgLoading && <p className="text-slate-400">正在加载配置...</p>}
        {!cfgLoading && !token && (
          <div className="mb-6">
            <Pairing />
          </div>
        )}

        {tab === "upload" && <Uploader />}
        {tab === "download" && <FileList />}
        {tab === "chat" && <Chat />}
        {tab === "devices" && <Devices />}
      </main>

      <footer className="sticky bottom-0 bg-slate-900/70 backdrop-blur border-t border-slate-800">
        <div className="max-w-5xl mx-auto px-4 py-3 text-sm text-slate-400">
          本地网络文件同步与聊天 · Tailwind · Next.js
        </div>
      </footer>
    </div>
  );
}
