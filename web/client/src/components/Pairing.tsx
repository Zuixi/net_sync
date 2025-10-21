"use client";
import { useState } from "react";
import { useAuth } from "@/lib/auth";

export default function Pairing() {
  const [tokenInput, setTokenInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { saveToken } = useAuth();

  async function handlePair() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/pair", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token: tokenInput.trim(),
          device_id: crypto.randomUUID(),
          device_name: navigator.userAgent.includes("Mobile") ? "Mobile Device" : "Desktop Browser",
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