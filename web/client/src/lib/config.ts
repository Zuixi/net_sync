"use client";
import { useEffect, useState } from "react";

export type ApiConfig = {
  server: { url: string; address: string; protocol: string };
  endpoints: { health: string; websocket: string; upload: string; api_base: string };
};

export function useApiConfig() {
  const [config, setConfig] = useState<ApiConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    fetch("/api/config")
      .then((r) => r.json())
      .then((data) => setConfig(data))
      .catch((e) => setError(e))
      .finally(() => setLoading(false));
  }, []);

  return { config, loading, error };
}