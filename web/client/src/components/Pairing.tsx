"use client";
import { useState } from "react";
import { useAuth } from "@/lib/auth";
import { DEVICE_CONFIG } from "@/lib/config";

export default function Pairing() {
  const [tokenInput, setTokenInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { saveToken } = useAuth();

  async function handlePair() {
    setLoading(true);
    setError(null);
    try {
      // 使用与 WebSocket 连接时一致的设备名
      const deviceName = navigator.userAgent.includes("Mobile") ? "移动设备" : "网页浏览器";

      // 保存设备名到 localStorage，确保与 WebSocket 连接时使用的一致
      localStorage.setItem("easy_sync_device_name", deviceName);

      const res = await fetch("/api/pair", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token: tokenInput.trim(),
          device_id: 'device_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9),
          device_name: deviceName,
        }),
      });
      const data = await res.json();
      if (data.token) {
        saveToken(data.token);
      } else {
        setError("配对成功但未返回 token");
      }
    } catch (e: any) {
      setError(e?.message || "配对失败");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
      <h3 className="text-lg font-semibold mb-2">手动配对</h3>
      <p className="text-sm text-slate-400 mb-4">
        配对令牌可在服务器数据目录的 <code className="bg-slate-800 px-1 py-0.5 rounded">pairing-token.txt</code> 文件中找到
      </p>
      <div className="flex gap-2">
        <input
          value={tokenInput}
          onChange={(e) => setTokenInput(e.target.value)}
          placeholder="输入配对令牌"
          className="flex-1 rounded-md bg-slate-800 px-3 py-2 text-sm outline-none focus:ring-2 ring-sky-600"
        />
        <button
          onClick={handlePair}
          disabled={loading}
          className="rounded-md bg-sky-600 px-3 py-2 text-sm hover:bg-sky-500 disabled:opacity-50"
        >
          {loading ? "正在配对..." : "连接"}
        </button>
      </div>
      {error && <p className="mt-2 text-sm text-rose-400">{error}</p>}
    </div>
  );
}