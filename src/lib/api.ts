import type { Task, TaskList } from "../types";
import { createAdapter } from './adapter/factory';
import type { DatabaseAdapter } from './adapter/types';
import type { Task as AdapterTask, TaskList as AdapterTaskList } from './adapter/types';

const adapter: DatabaseAdapter = createAdapter();

// Environment detection — the Go backend sets window.__GO_DESKTOP__.
export function isDesktop(): boolean {
  try {
    return !!(window as any).__GO_DESKTOP__;
  } catch {
    return false;
  }
}

// ==================== Type mapping helpers ====================

function adapterTaskToApp(a: AdapterTask): Task {
  return {
    id: a.id,
    title: a.title,
    description: a.description,
    completed: a.completed,
    priority: a.priority as Task["priority"],
    status: a.status as Task["status"],
    listId: a.list_id,
    categoryId: undefined, // adapter doesn't have category_id
    tags: safeJsonParse(a.tags, []) as Task["tags"],
    subTasks: safeJsonParse(a.sub_tasks, []) as Task["subTasks"],
    reminders: safeJsonParse(a.reminders, []) as Task["reminders"],
    dueDate: a.due_date ? new Date(a.due_date) : undefined,
    createdAt: new Date(a.created_at),
    updatedAt: new Date(a.updated_at),
    order: a.order_num,
    deleted: a.deleted || false,
    deletedAt: a.deleted_at ? new Date(a.deleted_at) : undefined,
  };
}

function appTaskToAdapter(t: Task): AdapterTask {
  return {
    id: t.id,
    title: t.title,
    description: t.description,
    completed: t.completed,
    priority: t.priority,
    status: t.status,
    list_id: t.listId,
    tags: JSON.stringify(t.tags),
    sub_tasks: JSON.stringify(t.subTasks),
    reminders: JSON.stringify(t.reminders),
    due_date: t.dueDate instanceof Date ? t.dueDate.toISOString() : (t.dueDate ?? undefined),
    created_at: t.createdAt instanceof Date ? t.createdAt.toISOString() : String(t.createdAt),
    updated_at: t.updatedAt instanceof Date ? t.updatedAt.toISOString() : String(t.updatedAt),
    order_num: t.order,
    deleted: t.deleted || false,
    deleted_at: t.deletedAt instanceof Date ? t.deletedAt.toISOString() : (t.deletedAt ?? undefined),
  };
}

function adapterListToApp(l: AdapterTaskList): TaskList {
  return {
    id: l.id,
    name: l.name,
    icon: l.icon,
    color: l.color,
    order: l.order_num,
    createdAt: new Date(l.created_at),
  };
}

// Used by databaseApi.saveList when it is implemented later
function _appListToAdapter(l: TaskList): AdapterTaskList {
  return {
    id: l.id,
    name: l.name,
    icon: l.icon,
    color: l.color,
    order_num: l.order,
    created_at: l.createdAt instanceof Date ? l.createdAt.toISOString() : String(l.createdAt),
    updated_at: new Date().toISOString(),
  };
}
void _appListToAdapter; // suppress unused warning

function safeJsonParse(raw: string, fallback: unknown): unknown {
  try {
    return JSON.parse(raw || "[]");
  } catch {
    return fallback;
  }
}

// ==================== Database API ====================

export const databaseApi = {
  async initDatabase(): Promise<void> {
    try {
      return await adapter.init();
    } catch {
      console.log("[API] Adapter init failed, running without backend");
    }
  },

  async getTasks(): Promise<Task[]> {
    try {
      const tasks = await adapter.getTasks();
      return tasks.map(adapterTaskToApp);
    } catch {
      // Fallback to localStorage when Go backend is not available
      const stored = localStorage.getItem("tasks");
      return stored ? JSON.parse(stored) : [];
    }
  },

  async saveTask(task: Task): Promise<Task> {
    try {
      const adapterTask = appTaskToAdapter(task);
      const result = await adapter.saveTask(adapterTask);
      return adapterTaskToApp(result);
    } catch {
      // Fallback to localStorage
      const tasks = await this.getTasks();
      const index = tasks.findIndex((t) => t.id === task.id);
      if (index >= 0) {
        tasks[index] = task;
      } else {
        tasks.push(task);
      }
      localStorage.setItem("tasks", JSON.stringify(tasks));
      return task;
    }
  },

  async deleteTask(id: string): Promise<void> {
    try {
      return await adapter.deleteTask(id);
    } catch {
      // Fallback to localStorage
      const tasks = await this.getTasks();
      const filtered = tasks.filter((t) => t.id !== id);
      localStorage.setItem("tasks", JSON.stringify(filtered));
    }
  },

  async getLists(): Promise<TaskList[]> {
    try {
      const lists = await adapter.getLists();
      return lists.map(adapterListToApp);
    } catch {
      // Fallback to localStorage
      const stored = localStorage.getItem("lists");
      return stored ? JSON.parse(stored) : [];
    }
  },

  // Check and send due reminders
  async checkDueReminders(): Promise<void> {
    try {
      const res = await fetch("/api/v1/notifications/check-due-reminders", { method: "POST" });
      if (!res.ok) throw new Error(`checkDueReminders: ${res.status}`);
    } catch {
      // Web/PWA: not applicable
    }
  },
};

// ==================== Notification API ====================

// 通知设置类型
export interface NotificationSettings {
  enabled: boolean;
  wechat_webhook?: string;
}

export const notificationApi = {
  // 发送系统通知
  async sendNotification(title: string, body: string): Promise<void> {
    try {
      await fetch("/api/v1/notifications/send", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, body }),
      });
    } catch {
      // Fallback to browser Notification API
      if ("Notification" in window) {
        if (Notification.permission === "granted") {
          new Notification(title, { body });
        } else if (Notification.permission !== "denied") {
          Notification.requestPermission().then((permission) => {
            if (permission === "granted") {
              new Notification(title, { body });
            }
          });
        }
      }
    }
  },

  // 发送企业微信机器人通知
  async sendWechatNotification(webhookUrl: string, title: string, content: string): Promise<void> {
    try {
      await fetch("/api/v1/notifications/wechat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ webhookUrl, title, content }),
      });
    } catch {
      // Web 端直接使用 fetch
      const response = await fetch(webhookUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          msgtype: "text",
          text: {
            content: `${title}\n\n${content}`,
          },
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to send wechat notification: ${response.statusText}`);
      }
    }
  },

  // 获取需要提醒的任务
  async getDueReminders(): Promise<any[]> {
    try {
      const res = await fetch("/api/v1/notifications/due-reminders");
      if (!res.ok) throw new Error(`getDueReminders: ${res.status}`);
      return res.json();
    } catch {
      return [];
    }
  },

  // 保存通知设置
  async saveSettings(settings: NotificationSettings): Promise<void> {
    try {
      await fetch("/api/v1/notifications/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(settings),
      });
    } catch {
      localStorage.setItem("notification_settings", JSON.stringify(settings));
    }
  },

  // 加载通知设置
  async loadSettings(): Promise<NotificationSettings | null> {
    try {
      const res = await fetch("/api/v1/notifications/settings");
      if (!res.ok) throw new Error(`loadSettings: ${res.status}`);
      return res.json();
    } catch {
      const stored = localStorage.getItem("notification_settings");
      return stored ? JSON.parse(stored) : null;
    }
  },

  // 请求通知权限
  async requestPermission(): Promise<boolean> {
    try {
      const res = await fetch("/api/v1/notifications/request-permission", { method: "POST" });
      if (!res.ok) throw new Error(`requestPermission: ${res.status}`);
      const data = await res.json();
      return data.granted === true;
    } catch {
      if ("Notification" in window) {
        const permission = await Notification.requestPermission();
        return permission === "granted";
      }
      return false;
    }
  },

  // 检查通知权限状态
  async checkPermission(): Promise<"granted" | "denied" | "default"> {
    try {
      const res = await fetch("/api/v1/notifications/check-permission");
      if (!res.ok) throw new Error(`checkPermission: ${res.status}`);
      const data = await res.json();
      return data.status as "granted" | "denied" | "default";
    } catch {
      if ("Notification" in window) {
        return Notification.permission as "granted" | "denied" | "default";
      }
      return "default";
    }
  },
};

// ==================== WebDAV API ====================

// 备份设置
export interface WebDavSettings {
  enabled: boolean;
  url: string;
  username: string;
  password: string;
  basePath: string;
  autoBackup: boolean;
  maxBackups?: number;
  simpleMode: boolean;
}

export const webdavApi = {
  async saveSettings(settings: WebDavSettings): Promise<void> {
    try {
      await fetch("/api/v1/webdav/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(settings),
      });
    } catch {
      localStorage.setItem("webdav_settings", JSON.stringify(settings));
    }
  },

  async loadSettings(): Promise<WebDavSettings | null> {
    try {
      const res = await fetch("/api/v1/webdav/settings");
      if (!res.ok) throw new Error(`loadSettings: ${res.status}`);
      return res.json();
    } catch {
      const s = localStorage.getItem("webdav_settings");
      return s ? JSON.parse(s) : null;
    }
  },

  async testConnection(settings: WebDavSettings): Promise<void> {
    try {
      await fetch("/api/v1/webdav/test-connection", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(settings),
      });
    } catch {
      await new Promise((r) => setTimeout(r, 500));
    }
  },

  async backup(): Promise<string> {
    try {
      const res = await fetch("/api/v1/webdav/backup", { method: "POST" });
      if (!res.ok) throw new Error(`backup: ${res.status}`);
      const data = await res.json();
      return data.filename || "web-backup";
    } catch {
      return "web-backup-placeholder";
    }
  },

  async restore(filename: string): Promise<void> {
    try {
      await fetch("/api/v1/webdav/restore", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ filename }),
      });
    } catch {
      // No-op when Go backend is not available
    }
  },
};
