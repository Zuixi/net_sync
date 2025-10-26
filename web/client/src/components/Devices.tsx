"use client";
import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth";
import { STORAGE_CONFIG } from "@/lib/config";

export default function Devices() {
  const { token } = useAuth();
  const [devices, setDevices] = useState<any[]>([]);
  const [currentDeviceId, setCurrentDeviceId] = useState<string | null>(null);

  // 从 localStorage 读取当前设备 ID
  useEffect(() => {
    const deviceId = localStorage.getItem(STORAGE_CONFIG.DEVICE_ID_KEY);
    setCurrentDeviceId(deviceId);
  }, []);

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
        {devices.map((d, i) => {
          // 判断是否为当前设备（通过设备 ID 匹配）
          const currentDeviceId = localStorage.getItem(STORAGE_CONFIG.DEVICE_ID_KEY);
          const isCurrentDevice = d.id === currentDeviceId;

          return (
            <div
              key={i}
              className={`rounded-md px-3 py-2 ${
                isCurrentDevice
                  ? "bg-sky-900/30 border border-sky-700"
                  : "bg-slate-800"
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="text-sm font-medium">{d.device_name || d.id}</div>
                {isCurrentDevice && (
                  <span className="text-xs text-sky-400 bg-sky-900/50 px-2 py-0.5 rounded">
                    当前设备
                  </span>
                )}
              </div>
              <div className="text-xs text-slate-400 mt-1">
                {d.last_ping
                  ? `活跃: ${new Date(d.last_ping).toLocaleTimeString()}`
                  : "活跃状态未知"}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}