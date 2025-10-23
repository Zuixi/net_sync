"use client";
import { useEffect, useState } from "react";
import { useAuth } from "./auth";

/**
 * 自动配对Hook
 * 在应用启动时自动获取pairing token并完成配对
 */
export function useAutoPair() {
  const { token, saveToken } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [autoPaired, setAutoPaired] = useState(false);
  const [attempted, setAttempted] = useState(false);

  useEffect(() => {
    // 如果已有token,标记为已配对
    if (token) {
      setAutoPaired(true);
      setAttempted(true);
      return;
    }

    // 如果已经尝试过配对,不重复执行
    if (attempted) return;

    // 执行自动配对
    async function autoPair() {
      setLoading(true);
      setError(null);
      setAttempted(true);

      try {
        console.log("开始自动配对...");

        // 1. 获取pairing token
        console.log("正在获取pairing token...");
        const tokenRes = await fetch("/api/pairing-token");
        if (!tokenRes.ok) {
          throw new Error(`无法获取配对令牌: ${tokenRes.status} ${tokenRes.statusText}`);
        }
        const tokenData = await tokenRes.json();
        const pairingToken = tokenData.token;
        console.log("成功获取pairing token:", pairingToken);

        // 2. 生成设备信息
        const deviceId = `device_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        const deviceName = navigator.userAgent.includes("Mobile")
          ? "移动设备"
          : "网页浏览器";
        console.log("设备信息:", { deviceId, deviceName });

        // 3. 执行配对
        console.log("正在执行配对...");
        const pairRes = await fetch("/api/pair", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            token: pairingToken,
            device_id: deviceId,
            device_name: deviceName,
          }),
        });

        if (!pairRes.ok) {
          const errorData = await pairRes.json().catch(() => ({}));
          throw new Error(errorData.error || `配对失败: ${pairRes.status}`);
        }

        const pairData = await pairRes.json();
        console.log("配对响应:", pairData);

        if (pairData.token) {
          saveToken(pairData.token);
          setAutoPaired(true);
          console.log("自动配对成功!");
        } else {
          throw new Error("配对成功但未返回token");
        }
      } catch (e: any) {
        console.error("自动配对失败:", e);
        setError(e?.message || "自动配对失败");
      } finally {
        setLoading(false);
      }
    }

    autoPair();
  }, [token, saveToken, attempted]);

  return { loading, error, autoPaired };
}
