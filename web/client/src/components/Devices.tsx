"use client";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";

export default function Devices() {
  const { token } = useAuth();
  const [devices, setDevices] = useState<any[]>([]);

  async function loadDevices() {
    if (!token) return;
    const res = await fetch("/api/devices", { headers: { Authorization: `Bearer ${token}` } });
    const data = await res.json();
    setDevices(data.devices || []);
  }

  useEffect(() => { loadDevices(); }, [token]);

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm text-slate-300">已发现设备</h2>
        <button onClick={loadDevices} className="rounded-md bg-slate-800 px-3 py-1.5 text-xs hover:bg-slate-700">刷新</button>
      </div>
      <div className="mt-3 space-y-2">
        {devices.length === 0 && <p className="text-sm text-slate-400">暂无设备</p>}
        {devices.map((d, i) => (
          <div key={i} className="rounded-md bg-slate-800 px-3 py-2">
            <div className="text-sm font-medium">{d.device_name || d.id}</div>
            <div className="text-xs text-slate-400">{d.address || "未知地址"}</div>
          </div>
        ))}
      </div>
    </div>
  );
}