import type { DatabaseAdapter, Task, TaskList, Category, Tag, ChangeRecord } from './types';

const BASE = ''; // 同源, 用相对路径

export class GoDesktopAdapter implements DatabaseAdapter {
  async init(): Promise<void> {
    // Go 服务器在启动时已初始化数据库
  }

  async getTasks(): Promise<Task[]> {
    const res = await fetch(`${BASE}/api/v1/tasks`);
    if (!res.ok) throw new Error(`getTasks: ${res.status}`);
    return res.json();
  }

  async saveTask(task: Task): Promise<Task> {
    const res = await fetch(`${BASE}/api/v1/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    });
    if (!res.ok) throw new Error(`saveTask: ${res.status}`);
    return res.json();
  }

  async deleteTask(id: string): Promise<void> {
    const res = await fetch(`${BASE}/api/v1/tasks/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error(`deleteTask: ${res.status}`);
  }

  async getLists(): Promise<TaskList[]> {
    const res = await fetch(`${BASE}/api/v1/lists`);
    if (!res.ok) throw new Error(`getLists: ${res.status}`);
    return res.json();
  }

  async saveList(list: TaskList): Promise<TaskList> {
    const res = await fetch(`${BASE}/api/v1/lists`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(list),
    });
    if (!res.ok) throw new Error(`saveList: ${res.status}`);
    return res.json();
  }

  async deleteList(id: string): Promise<void> {
    const res = await fetch(`${BASE}/api/v1/lists/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error(`deleteList: ${res.status}`);
  }

  async getCategories(): Promise<Category[]> {
    const res = await fetch(`${BASE}/api/v1/categories`);
    if (!res.ok) throw new Error(`getCategories: ${res.status}`);
    return res.json();
  }

  async saveCategory(cat: Category): Promise<Category> {
    const res = await fetch(`${BASE}/api/v1/categories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cat),
    });
    if (!res.ok) throw new Error(`saveCategory: ${res.status}`);
    return res.json();
  }

  async deleteCategory(id: string): Promise<void> {
    const res = await fetch(`${BASE}/api/v1/categories/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error(`deleteCategory: ${res.status}`);
  }

  async getTags(): Promise<Tag[]> {
    const res = await fetch(`${BASE}/api/v1/tags`);
    if (!res.ok) throw new Error(`getTags: ${res.status}`);
    return res.json();
  }

  async saveTag(tag: Tag): Promise<Tag> {
    const res = await fetch(`${BASE}/api/v1/tags`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(tag),
    });
    if (!res.ok) throw new Error(`saveTag: ${res.status}`);
    return res.json();
  }

  async deleteTag(id: string): Promise<void> {
    const res = await fetch(`${BASE}/api/v1/tags/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error(`deleteTag: ${res.status}`);
  }

  async getUnsyncedChanges(): Promise<ChangeRecord[]> {
    throw new Error("getUnsyncedChanges: not implemented — desktop sync is handled by Go SyncEngine");
  }

  async markChangesSynced(_changes: ChangeRecord[]): Promise<void> {
    throw new Error("markChangesSynced: not implemented — desktop sync is handled by Go SyncEngine");
  }
}
