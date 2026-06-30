import { SyncClient, type SyncRecord } from './client';
import type { DatabaseAdapter } from '../adapter/types';

export class SyncEngine {
  private client: SyncClient;
  private lastSync: string;
  private pushTimer: number | null = null;
  private pullTimer: number | null = null;

  constructor(
    private adapter: DatabaseAdapter,
    serverUrl: string,
    token: string,
  ) {
    this.client = new SyncClient(serverUrl, token);
    this.lastSync = new Date().toISOString();
  }

  start(): void {
    // 启动时立即拉取
    this.doPull();

    // 每 3 秒推送
    this.pushTimer = window.setInterval(() => this.doPush(), 3000);
    // 每 30 秒拉取
    this.pullTimer = window.setInterval(() => this.doPull(), 30000);
  }

  stop(): void {
    if (this.pushTimer) clearInterval(this.pushTimer);
    if (this.pullTimer) clearInterval(this.pullTimer);
  }

  private async doPush(): Promise<void> {
    try {
      const changes = await (this.adapter as any).getUnsyncedChanges?.();
      if (!changes || changes.length === 0) return;

      const records: SyncRecord[] = changes.map((c: any) => ({
        table_name: c.table_name,
        record_id: c.record_id,
        action: c.action,
        timestamp: c.timestamp,
      }));

      await this.client.push(records);
      await (this.adapter as any).markChangesSynced?.(changes);
    } catch (err) {
      console.warn('[sync] push failed:', err);
    }
  }

  private async doPull(): Promise<void> {
    try {
      const resp = await this.client.pull(this.lastSync);
      this.lastSync = resp.server_time;
      if (resp.changes.length > 0) {
        // TODO Phase 4: LWW merge
        console.log('[sync] pulled', resp.changes.length, 'changes');
      }
    } catch (err) {
      console.warn('[sync] pull failed:', err);
    }
  }
}
