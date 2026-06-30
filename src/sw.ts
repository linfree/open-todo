// Service Worker — compiled by vite-plugin-pwa (injectManifest)
// Type-checked separately; excluded from the main tsconfig.
/// <reference lib="webworker" />

import { precacheAndRoute } from "workbox-precaching";
import { registerRoute } from "workbox-routing";
import { NetworkFirst, CacheFirst } from "workbox-strategies";
import { ExpirationPlugin } from "workbox-expiration";

declare const self: ServiceWorkerGlobalScope;

// ── Precaching (manifest injected by vite-plugin-pwa at build time) ──

precacheAndRoute(self.__WB_MANIFEST);

// ── Runtime caching (mirrors the generateSW config from Task 13) ──

registerRoute(
  /\/api\/v1\/.*/,
  new NetworkFirst({
    cacheName: "api-cache",
    plugins: [
      new ExpirationPlugin({ maxEntries: 100, maxAgeSeconds: 60 * 60 }),
    ],
  }),
);

registerRoute(
  /\.(js|css|png|jpg|svg|ico|woff2?)$/,
  new CacheFirst({
    cacheName: "static-cache",
    plugins: [
      new ExpirationPlugin({ maxEntries: 200, maxAgeSeconds: 30 * 24 * 3600 }),
    ],
  }),
);

// ── Lifecycle events ──

self.addEventListener("install", () => {
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

// ── Background sync ──

self.addEventListener("sync", (event) => {
  if (event.tag === "sync-changes") {
    event.waitUntil(syncChanges());
  }
});

// ── IndexedDB helpers ──

const DB_NAME = "open-todo";
const STORE_NAME = "syncLog";

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

function getAllFromIndex<T>(
  store: IDBObjectStore,
  indexName: string,
  value: IDBValidKey,
): Promise<T[]> {
  return new Promise((resolve, reject) => {
    const index = store.index(indexName);
    const req = index.getAll(value);
    req.onsuccess = () => resolve(req.result as T[]);
    req.onerror = () => reject(req.error);
  });
}

function transactionComplete(tx: IDBTransaction): Promise<void> {
  return new Promise((resolve, reject) => {
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
    tx.onabort = () => reject(tx.error);
  });
}

// ── Sync logic ──

interface SyncLogRecord {
  id?: number;
  table_name: string;
  record_id: string;
  action: string;
  timestamp: string;
  synced: boolean;
}

async function syncChanges(): Promise<void> {
  let db: IDBDatabase | null = null;
  try {
    db = await openDB();

    // Read unsynced records (synced === 0 / false)
    const readTx = db.transaction(STORE_NAME, "readonly");
    const store = readTx.objectStore(STORE_NAME);
    const unsynced = await getAllFromIndex<SyncLogRecord>(
      store,
      "synced",
      0,
    );
    await transactionComplete(readTx);

    if (unsynced.length === 0) return;

    const changes = unsynced.map((r) => ({
      table_name: r.table_name,
      record_id: r.record_id,
      action: r.action,
      timestamp: r.timestamp,
    }));

    // Push to server
    const res = await fetch("/api/v1/sync/push", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ changes }),
    });

    if (!res.ok) {
      console.error("[SW] Sync push failed:", res.status);
      return;
    }

    // Mark unsynced records as synced
    const writeTx = db.transaction(STORE_NAME, "readwrite");
    const writeStore = writeTx.objectStore(STORE_NAME);
    for (const record of unsynced) {
      writeStore.put({ ...record, synced: true });
    }
    await transactionComplete(writeTx);
  } catch (err) {
    console.error("[SW] Background sync error:", err);
  } finally {
    if (db) db.close();
  }
}
