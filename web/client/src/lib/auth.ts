"use client";
import { useEffect, useState } from "react";

const KEY = "easy_sync_token";

export function useAuth() {
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    const t = localStorage.getItem(KEY);
    if (t) setToken(t);
  }, []);

  function saveToken(t: string) {
    localStorage.setItem(KEY, t);
    setToken(t);
  }

  function clearToken() {
    localStorage.removeItem(KEY);
    setToken(null);
  }

  return { token, saveToken, clearToken };
}