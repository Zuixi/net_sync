"use client";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";

function formatSize(n: number) {
  const units = ["B", "KB", "MB", "GB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(1)} ${units[i]}`;
}

export default function FileList() {
  const { token } = useAuth();
  const [files, setFiles] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  async function loadFiles() {
    if (!token) return;
    setLoading(true);
    try {
      const res = await fetch("/api/files", { headers: { Authorization: `Bearer ${token}` } });
      const data = await res.json();
      setFiles(data.files || []);
    } catch (e) {
      console.error("Load files failed", e);
    } finally {
      setLoading(false);
    }
  }

  async function del(id: string) {
    if (!token) return;
    await fetch(`/api/files/${id}`, { method: "DELETE", headers: { Authorization: `Bearer ${token}` } });
    await loadFiles();
  }

  function download(id: string) {
    window.open(`/files/${id}`, "_blank");
  }

  async function verify(id: string) {
    const res = await fetch(`/files/${id}/sha256`);
    const data = await res.json();
    alert(`SHA-256: ${data.sha256 || "unknown"}`);
  }

  useEffect(() => { loadFiles(); }, [token]);

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm text-slate-300">可下载文件</h2>
        <button onClick={loadFiles} className="rounded-md bg-slate-800 px-3 py-1.5 text-xs hover:bg-slate-700">刷新</button>
      </div>
      <div className="mt-3 space-y-2">
        {loading && <p className="text-sm text-slate-400">加载中...</p>}
        {!loading && files.length === 0 && <p className="text-sm text-slate-400">暂无文件</p>}
        {!loading && files.map((f) => (
          <div key={f.id} className="flex items-center justify-between rounded-md bg-slate-800 px-3 py-2">
            <div>
              <div className="text-sm font-medium">{f.name || f.id}</div>
              <div className="text-xs text-slate-400">{formatSize(f.size || 0)}</div>
            </div>
            <div className="flex gap-2">
              <button onClick={() => verify(f.id)} className="rounded-md bg-slate-700 px-2.5 py-1.5 text-xs hover:bg-slate-600">校验</button>
              <button onClick={() => download(f.id)} className="rounded-md bg-sky-600 px-2.5 py-1.5 text-xs hover:bg-sky-500">下载</button>
              <button onClick={() => del(f.id)} className="rounded-md bg-rose-600 px-2.5 py-1.5 text-xs hover:bg-rose-500">删除</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}