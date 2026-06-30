import type { DatabaseAdapter, Task, TaskList } from './types';

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
}
