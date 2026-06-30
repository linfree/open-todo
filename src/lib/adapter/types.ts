export interface Task {
  id: string;
  user_id?: string;
  title: string;
  description?: string;
  completed: boolean;
  priority: string;
  status: string;
  list_id: string;
  tags: string;
  sub_tasks: string;
  reminders: string;
  due_date?: string;
  deleted: boolean;
  deleted_at?: string;
  order_num: number;
  created_at: string;
  updated_at: string;
}

export interface TaskList {
  id: string;
  user_id?: string;
  name: string;
  icon?: string;
  color?: string;
  order_num: number;
  created_at: string;
  updated_at: string;
}

export interface ChangeRecord {
  id: number;
  table_name: string;
  record_id: string;
  action: string;
  timestamp: string;
  synced: boolean;
}

export interface DatabaseAdapter {
  init(): Promise<void>;
  getTasks(): Promise<Task[]>;
  saveTask(task: Task): Promise<Task>;
  deleteTask(id: string): Promise<void>;
  getLists(): Promise<TaskList[]>;
  saveList(list: TaskList): Promise<TaskList>;
  deleteList(id: string): Promise<void>;
  getUnsyncedChanges(): Promise<ChangeRecord[]>;
  markChangesSynced(changes: ChangeRecord[]): Promise<void>;
}
