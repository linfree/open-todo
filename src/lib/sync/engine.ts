import { SyncClient, type SyncRecord } from './client';
import type { DatabaseAdapter, Task, TaskList } from '../adapter/types';

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
    this.doPull();
    this.pushTimer = window.setInterval(() => this.doPush(), 3000);
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
        data: c.data,
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

      let merged = 0;
      for (const c of resp.changes) {
        const existing = await this.getLocalRecord(c.table_name, c.record_id);
        let shouldWrite = !existing;

        if (!shouldWrite && existing && c.data) {
          const remoteTime = new Date(c.timestamp).getTime();
          const localTime = new Date(existing.updated_at).getTime();
          if (!isNaN(remoteTime) && !isNaN(localTime) && remoteTime >= localTime) {
            shouldWrite = true;
          }
        }

        if (shouldWrite && c.data) {
          await this.saveRemoteRecord(c.table_name, c.data);
          merged++;
        }
      }

      if (merged > 0) {
        console.log('[sync] merged', merged, 'of', resp.changes.length, 'changes');
      }
    } catch (err) {
      console.warn('[sync] pull failed:', err);
    }
  }

  private async getLocalRecord(table: string, id: string): Promise<{ updated_at: string } | null> {
    try {
      if (table === 'tasks') {
        const tasks = await this.adapter.getTasks();
        const task = tasks.find((t: Task) => t.id === id);
        return task ? { updated_at: task.updated_at } : null;
      }
      if (table === 'lists') {
        const lists = await this.adapter.getLists();
        const list = lists.find((l: TaskList) => l.id === id);
        return list ? { updated_at: list.updated_at } : null;
      }
    } catch { /* offline or error */ }
    return null;
  }

  private async saveRemoteRecord(table: string, data: any): Promise<void> {
    if (table === 'tasks') {
      await this.adapter.saveTask(data as Task);
    } else if (table === 'lists') {
      await this.adapter.saveList(data as TaskList);
    }
  }
}
