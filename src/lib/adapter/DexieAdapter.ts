import Dexie, { type Table } from 'dexie';
import type { DatabaseAdapter, Task, TaskList, ChangeRecord } from './types';
import { v4 as uuid } from './uuid';

class TodoDB extends Dexie {
  tasks!: Table<Task>;
  lists!: Table<TaskList>;
  syncLog!: Table<ChangeRecord>;

  constructor() {
    super('open-todo');
    this.version(1).stores({
      tasks: 'id, user_id, list_id, updated_at, [user_id+list_id]',
      lists: 'id, user_id, updated_at',
      syncLog: '++id, table_name, record_id, synced, timestamp',
    });
  }
}

const db = new TodoDB();

function now(): string {
  return new Date().toISOString();
}

export class DexieAdapter implements DatabaseAdapter {
  async init(): Promise<void> {
    const lists = await db.lists.count();
    if (lists === 0) {
      const ts = now();
      await db.lists.bulkPut([
        { id: 'all', name: '全部', icon: 'Inbox', order_num: 0, created_at: ts, updated_at: ts },
        { id: 'today', name: '今天', icon: 'Sun', order_num: 1, created_at: ts, updated_at: ts },
        { id: 'week', name: '最近7天', icon: 'Calendar', order_num: 2, created_at: ts, updated_at: ts },
      ]);
    }
  }

  async getTasks(): Promise<Task[]> {
    return db.tasks.filter(t => !t.deleted).toArray();
  }

  async saveTask(task: Task): Promise<Task> {
    const t = {
      ...task,
      id: task.id || uuid(),
      created_at: task.created_at || now(),
      updated_at: now(),
    };
    await db.tasks.put(t);
    await db.syncLog.put({
      table_name: 'tasks',
      record_id: t.id,
      action: 'insert',
      timestamp: t.updated_at,
      synced: false,
    } as ChangeRecord);
    return t;
  }

  async deleteTask(id: string): Promise<void> {
    const ts = now();
    await db.tasks.update(id, { deleted: true, deleted_at: ts, updated_at: ts });
    await db.syncLog.put({
      table_name: 'tasks',
      record_id: id,
      action: 'delete',
      timestamp: ts,
      synced: false,
    } as ChangeRecord);
  }

  async getLists(): Promise<TaskList[]> {
    return db.lists.toArray();
  }

  async saveList(list: TaskList): Promise<TaskList> {
    const l = {
      ...list,
      id: list.id || uuid(),
      created_at: list.created_at || now(),
      updated_at: now(),
    };
    await db.lists.put(l);
    return l;
  }

  async deleteList(id: string): Promise<void> {
    await db.lists.delete(id);
  }

  async getUnsyncedChanges(): Promise<ChangeRecord[]> {
    return db.syncLog.filter(c => !c.synced).toArray();
  }

  async markChangesSynced(changes: ChangeRecord[]): Promise<void> {
    const ids = changes.map(c => c.id);
    await db.syncLog.where('id').anyOf(ids).modify({ synced: true });
  }
}
