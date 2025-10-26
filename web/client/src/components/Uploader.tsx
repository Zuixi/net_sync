"use client";
import { useRef, useState } from "react";
import * as tus from "tus-js-client";
import { useAuth } from "@/lib/auth";
import { UPLOAD_CONFIG } from "@/lib/config";

function formatSize(n: number) {
  const units = ["B", "KB", "MB", "GB"];
  let i = 0;
  let v = n;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(1)} ${units[i]}`;
}

export default function Uploader() {
  const { token } = useAuth();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [logs, setLogs] = useState<string[]>([]);

  function appendLog(s: string) { setLogs((l) => [s, ...l].slice(0, UPLOAD_CONFIG.LOG_LIMIT)); }

  function pickFiles() { inputRef.current?.click(); }

  function startUpload(files: FileList | null) {
    if (!files || !token) return;
    Array.from(files).forEach((file) => {
      const upload = new tus.Upload(file, {
        endpoint: UPLOAD_CONFIG.TUS_ENDPOINT,
        retryDelays: UPLOAD_CONFIG.TUS_RETRY_DELAYS,
        metadata: { filename: file.name, filetype: file.type },
        headers: { Authorization: `Bearer ${token}` },
        onError: (err) => appendLog(`上传失败: ${err}`),
        onProgress: (uploaded, total) => {
          const pct = ((uploaded / total) * 100).toFixed(2);
          appendLog(`${file.name}: ${pct}%`);
        },
        onSuccess: () => {
          const id = (upload.url?.split("/").pop() || "").replace(/[^a-zA-Z0-9_-]/g, "");
          appendLog(`上传完成: ${file.name} (id=${id})`);
        },
      });
      upload.start();
    });
  }

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
      <div className="flex items-center gap-2">
        <button onClick={pickFiles} className="rounded-md bg-sky-600 px-3 py-2 text-sm hover:bg-sky-500">选择文件</button>
        <input type="file" multiple ref={inputRef} className="hidden" onChange={(e) => startUpload(e.target.files)} />
        {!token && <span className="text-sm text-slate-400">请先完成配对</span>}
      </div>
      <div className="mt-3 text-xs text-slate-400 space-y-1">
        {logs.map((l, i) => (<div key={i}>{l}</div>))}
      </div>
    </div>
  );
}