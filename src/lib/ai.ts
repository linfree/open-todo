import { useState, useEffect, useCallback } from "react";

interface AIStatus {
  enabled: boolean;
  configured: boolean;
  model: string;
}

let cachedStatus: AIStatus | null = null;
let lastFetchTime = 0;
const CACHE_TTL = 30000; // 30 seconds

export async function checkAIStatus(): Promise<{ enabled: boolean; configured: boolean }> {
  const now = Date.now();
  if (cachedStatus && now - lastFetchTime < CACHE_TTL) {
    return { enabled: cachedStatus.enabled, configured: cachedStatus.configured };
  }

  try {
    const res = await fetch("/api/v1/ai/status");
    if (!res.ok) throw new Error(`AI status: ${res.status}`);
    const data: AIStatus = await res.json();
    cachedStatus = data;
    lastFetchTime = now;
    return { enabled: data.enabled, configured: data.configured };
  } catch {
    // If backend is not available, return disabled
    return { enabled: false, configured: false };
  }
}

export function useAIStatus() {
  const [status, setStatus] = useState<{ enabled: boolean; configured: boolean }>({
    enabled: false,
    configured: false,
  });
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    const s = await checkAIStatus();
    setStatus(s);
    setLoading(false);
  }, []);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, CACHE_TTL);
    return () => clearInterval(interval);
  }, [refresh]);

  return { ...status, loading, refresh };
}
