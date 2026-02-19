import { useEffect, useRef, useCallback } from "react";

export interface SSEMessage {
  type:
    | "connected"
    | "odds_updated"
    | "event_created"
    | "event_resolved"
    | "user_resolved"
    | "bingo_resolved"
    | "bingo_winner"
    | "activity_new";
  data?: Record<string, unknown>;
}

type Listener = (msg: SSEMessage) => void;

let globalSource: EventSource | null = null;
let listeners = new Set<Listener>();
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

function connect() {
  if (globalSource) return;

  const token = localStorage.getItem("token");
  const url = token ? `/api/stream?token=${encodeURIComponent(token)}` : "/api/stream";
  const source = new EventSource(url);
  globalSource = source;

  source.onmessage = (e) => {
    try {
      const msg: SSEMessage = JSON.parse(e.data);
      listeners.forEach((fn) => fn(msg));
    } catch {
      // ignore parse errors
    }
  };

  source.onerror = () => {
    source.close();
    globalSource = null;
    // Reconnect after 3s
    if (!reconnectTimer) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        if (listeners.size > 0) connect();
      }, 3000);
    }
  };
}

// Reconnect when tab becomes visible again (catches missed events)
if (typeof document !== "undefined") {
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible" && listeners.size > 0) {
      // Force reconnect to get a fresh stream
      if (globalSource) {
        globalSource.close();
        globalSource = null;
      }
      connect();
      // Notify listeners so pages can refetch stale data
      listeners.forEach((fn) => fn({ type: "connected" }));
    }
  });
}

function disconnect() {
  if (globalSource) {
    globalSource.close();
    globalSource = null;
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

/**
 * Subscribe to the global SSE stream. The connection is shared
 * across all components and stays alive as long as at least one
 * subscriber exists.
 */
export function useEventStream(onMessage: Listener) {
  const callbackRef = useRef(onMessage);
  callbackRef.current = onMessage;

  const stableListener = useCallback((msg: SSEMessage) => {
    callbackRef.current(msg);
  }, []);

  useEffect(() => {
    listeners.add(stableListener);
    connect();

    return () => {
      listeners.delete(stableListener);
      if (listeners.size === 0) {
        disconnect();
      }
    };
  }, [stableListener]);
}
