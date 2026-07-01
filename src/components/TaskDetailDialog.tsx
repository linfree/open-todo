import { useState, useEffect } from "react";
import { Tag, Plus, Trash2, Bell, Sparkles, Loader2, CheckCircle2, Circle, Clock, ListChecks, Flag, FolderOpen } from "lucide-react";
import { useTodoStore } from "../store/todoStore";
import { Priority, TaskStatus, Task, SubTask, Reminder, Tag as TagType } from "../types";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Textarea } from "./ui/textarea";
import { Select } from "./ui/select";
import { Checkbox } from "./ui/checkbox";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose } from "./ui/dialog";
import { ConfirmDialog } from "./ui/confirm-dialog";
import { DatePicker } from "./ui/date-picker";
import { Badge } from "./ui/badge";
import { cn } from "../lib/utils";
import { useAIStatus } from "../lib/ai";
import { TAG_COLORS } from "../lib/icons";

interface TaskDetailDialogProps {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
}

export function TaskDetailDialog({ task, isOpen, onClose }: TaskDetailDialogProps) {
  const { updateTask, permanentlyDeleteTask, toggleTaskComplete, categories, tags: storeTags, addTag } = useTodoStore();

  const [editing, setEditing] = useState(false);

  // 表单状态
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState<Priority>(Priority.NONE);
  const [status, setStatus] = useState<TaskStatus>(TaskStatus.TODO);
  const [listId, setListId] = useState("");
  const [categoryId, setCategoryId] = useState<string | undefined>(undefined);
  const [dueDate, setDueDate] = useState<Date | null>(null);
  const [selectedTags, setSelectedTags] = useState<TagType[]>([]);
  const [subTasks, setSubTasks] = useState<SubTask[]>([]);
  const [newSubTaskTitle, setNewSubTaskTitle] = useState("");
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [isAddingTag, setIsAddingTag] = useState(false);
  const [newTagName, setNewTagName] = useState("");

  // 删除确认
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<{ type: "task" } | { type: "subtask"; id: string } | { type: "reminder"; id: string } | null>(null);

  // AI breakdown state
  const { enabled: aiEnabled, configured: aiConfigured } = useAIStatus();
  const [aiBreakdownLoading, setAIBreakdownLoading] = useState(false);
  const [aiBreakdownError, setAIBreakdownError] = useState<string | null>(null);

  // 初始化表单（任务变化或对话框打开时，重置为查看模式）
  useEffect(() => {
    if (task) {
      setTitle(task.title);
      setDescription(task.description || "");
      setPriority(task.priority);
      setStatus(task.status);
      setListId(task.listId);
      setCategoryId(task.categoryId);
      setDueDate(task.dueDate ? new Date(task.dueDate) : null);
      setSelectedTags(task.tags || []);
      setSubTasks(task.subTasks || []);
      setReminders(task.reminders || []);
      setEditing(false);
    }
  }, [task]);

  const handleSave = () => {
    if (!task) return;

    updateTask(task.id, {
      title: title.trim(),
      description: description.trim() || undefined,
      priority,
      status,
      listId,
      categoryId,
      dueDate: dueDate || undefined,
      tags: selectedTags,
      subTasks,
      reminders,
    });

    setEditing(false);
  };

  const handleCancelEdit = () => {
    if (!task) return;
    // 丢弃修改，恢复到任务原始值
    setTitle(task.title);
    setDescription(task.description || "");
    setPriority(task.priority);
    setStatus(task.status);
    setListId(task.listId);
    setCategoryId(task.categoryId);
    setDueDate(task.dueDate ? new Date(task.dueDate) : null);
    setSelectedTags(task.tags || []);
    setSubTasks(task.subTasks || []);
    setReminders(task.reminders || []);
    setEditing(false);
  };

  const toggleTag = (tag: TagType) => {
    setSelectedTags((prev) =>
      prev.find((t) => t.id === tag.id)
        ? prev.filter((t) => t.id !== tag.id)
        : [...prev, tag]
    );
  };

  function handleAddTag() {
    if (!newTagName.trim()) return;

    const existingTag = storeTags.find((t) => t.name === newTagName.trim());
    if (existingTag) {
      if (!selectedTags.find((t) => t.id === existingTag.id)) {
        setSelectedTags((prev) => [...prev, existingTag]);
      }
    } else {
      const color = TAG_COLORS[Math.floor(Math.random() * TAG_COLORS.length)];
      const newTag: TagType = {
        id: crypto.randomUUID(),
        name: newTagName.trim(),
        color,
      };
      addTag(newTag);
      setSelectedTags((prev) => [...prev, newTag]);
    }

    setNewTagName("");
    setIsAddingTag(false);
  }

  const addSubTask = () => {
    if (!newSubTaskTitle.trim()) return;

    const newSubTask: SubTask = {
      id: crypto.randomUUID(),
      title: newSubTaskTitle.trim(),
      completed: false,
    };

    setSubTasks([...subTasks, newSubTask]);
    setNewSubTaskTitle("");
  };

  const handleAIBreakdown = async () => {
    setAIBreakdownLoading(true);
    setAIBreakdownError(null);

    try {
      const res = await fetch("/api/v1/ai/breakdown-task", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, description }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: `HTTP ${res.status}` }));
        throw new Error(data.error || `请求失败 (${res.status})`);
      }

      const result = await res.json();
      const newSubs: SubTask[] = (result.subtasks || []).map((st: { title: string }) => ({
        id: crypto.randomUUID(),
        title: st.title,
        completed: false,
      }));

      if (newSubs.length > 0) {
        setSubTasks([...subTasks, ...newSubs]);
      }
    } catch (err: any) {
      const msg = typeof err === "string" ? err : (err.message || "未知错误");
      setAIBreakdownError(msg);
    } finally {
      setAIBreakdownLoading(false);
    }
  };

  const toggleSubTask = (id: string) => {
    setSubTasks(
      subTasks.map((st) =>
        st.id === id ? { ...st, completed: !st.completed } : st
      )
    );
  };

  const deleteSubTask = (id: string) => {
    setDeleteTarget({ type: "subtask", id });
    setIsDeleteConfirmOpen(true);
  };

  const handleConfirmDelete = () => {
    if (!deleteTarget) return;
    if (deleteTarget.type === "task") {
      permanentlyDeleteTask(task!.id);
      onClose();
    } else if (deleteTarget.type === "subtask") {
      setSubTasks(subTasks.filter((st) => st.id !== deleteTarget.id));
    } else if (deleteTarget.type === "reminder") {
      setReminders(reminders.filter((r) => r.id !== deleteTarget.id));
    }
    setDeleteTarget(null);
  };

  const addReminder = () => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    tomorrow.setHours(9, 0, 0, 0);
    const newReminder: Reminder = {
      id: crypto.randomUUID(),
      date: tomorrow,
      repeat: "none",
      enabled: true,
    };
    setReminders([...reminders, newReminder]);
  };

  const updateReminder = (id: string, updates: Partial<Reminder>) => {
    setReminders(
      reminders.map((r) => (r.id === id ? { ...r, ...updates } : r))
    );
  };

  const deleteReminder = (id: string) => {
    setDeleteTarget({ type: "reminder", id });
    setIsDeleteConfirmOpen(true);
  };

  const priorityOptions = [
    { value: Priority.NONE, label: "无", color: "text-gray-500" },
    { value: Priority.LOW, label: "低", color: "text-blue-500" },
    { value: Priority.MEDIUM, label: "中", color: "text-yellow-500" },
    { value: Priority.HIGH, label: "高", color: "text-red-500" },
  ];

  const statusOptions = [
    { value: TaskStatus.TODO, label: "待办" },
    { value: TaskStatus.IN_PROGRESS, label: "进行中" },
    { value: TaskStatus.DONE, label: "已完成" },
  ];

  const priorityBadgeVariant = (p: Priority): "outline" | "info" | "warning" | "destructive" => {
    switch (p) {
      case Priority.LOW: return "info";
      case Priority.MEDIUM: return "warning";
      case Priority.HIGH: return "destructive";
      default: return "outline";
    }
  };

  const priorityLabel = (p: Priority) => {
    const opt = priorityOptions.find((o) => o.value === p);
    return opt?.label ?? "无";
  };

  const statusBadgeVariant = (s: TaskStatus): "secondary" | "info" | "success" => {
    switch (s) {
      case TaskStatus.TODO: return "secondary";
      case TaskStatus.IN_PROGRESS: return "info";
      case TaskStatus.DONE: return "success";
    }
  };

  const statusLabel = (s: TaskStatus) => {
    const opt = statusOptions.find((o) => o.value === s);
    return opt?.label ?? "待办";
  };

  const repeatLabel = (r: string) => {
    switch (r) {
      case "daily": return "每天";
      case "weekly": return "每周";
      case "monthly": return "每月";
      case "yearly": return "每年";
      default: return "不重复";
    }
  };

  const formatDate = (d: Date | string) => {
    const date = typeof d === "string" ? new Date(d) : d;
    return date.toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  if (!task) return null;

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{editing ? "编辑任务" : "任务详情"}</DialogTitle>
        </DialogHeader>

        <DialogClose onClick={onClose} />

        {/* 删除按钮（查看和编辑模式都显示） */}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => {
            setDeleteTarget({ type: "task" });
            setIsDeleteConfirmOpen(true);
          }}
          className="absolute top-3 right-12 h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
          title="删除任务"
        >
          <Trash2 className="w-4 h-4" />
        </Button>

        <div className="space-y-6">
          {/* 标题 */}
          <div>
            {editing ? (
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="任务标题"
                className="text-lg font-medium"
              />
            ) : (
              <h2 className="text-lg font-medium break-words">{title}</h2>
            )}
          </div>

          {/* 描述 */}
          <div>
            <label className="text-sm font-medium mb-2 block">描述</label>
            {editing ? (
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="添加描述..."
                className="min-h-[100px]"
              />
            ) : (
              <p className="text-sm text-muted-foreground whitespace-pre-wrap break-words min-h-[24px]">
                {description || "暂无描述"}
              </p>
            )}
          </div>

          {/* 优先级和状态 */}
          {editing ? (
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-2 block">优先级</label>
                <Select value={priority} onChange={(e) => setPriority(e.target.value as Priority)}>
                  {priorityOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <label className="text-sm font-medium mb-2 block">状态</label>
                <Select value={status} onChange={(e) => setStatus(e.target.value as TaskStatus)}>
                  {statusOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </div>
            </div>
          ) : (
            <div className="bg-muted/30 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-2">
                <Flag className="w-4 h-4 text-muted-foreground" />
                <span className="text-xs text-muted-foreground">优先级 / 状态</span>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="text-xs text-muted-foreground">优先级</span>
                  <div className="mt-1">
                    <Badge variant={priorityBadgeVariant(priority)}>
                      {priorityLabel(priority)}
                    </Badge>
                  </div>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground">状态</span>
                  <div className="mt-1">
                    <Badge variant={statusBadgeVariant(status)}>
                      {statusLabel(status)}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* 分类 */}
          {editing ? (
            <div>
              <label className="text-sm font-medium mb-2 block">分类</label>
              <Select value={categoryId || "inbox"} onChange={(e) => setCategoryId(e.target.value)}>
                {categories.map((cat) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.name}
                  </option>
                ))}
              </Select>
            </div>
          ) : (
            <div className="bg-muted/30 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <FolderOpen className="w-4 h-4 text-muted-foreground" />
                <span className="text-xs text-muted-foreground">分类</span>
              </div>
              <p className="text-sm">
                {categories.find((c) => c.id === categoryId)?.name || "收件箱"}
              </p>
            </div>
          )}

          {/* 截止时间 */}
          {editing ? (
            <div>
              <label className="text-sm font-medium mb-2 block">截止时间</label>
              <DatePicker
                value={dueDate}
                onChange={setDueDate}
                placeholder="选择截止日期和时间"
                showSeconds={true}
              />
            </div>
          ) : (
            <div className="bg-muted/30 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <Clock className="w-4 h-4 text-muted-foreground" />
                <span className="text-xs text-muted-foreground">截止时间</span>
              </div>
              {dueDate ? (
                <p className="text-sm font-medium">{formatDate(dueDate)}</p>
              ) : (
                <p className="text-sm text-muted-foreground">未设置</p>
              )}
            </div>
          )}

          {/* 标签 */}
          {editing ? (
            <div>
              <label className="text-sm font-medium mb-2 flex items-center gap-2">
                <Tag className="w-4 h-4 text-current" />
                标签
              </label>
              <>
                <div className="flex flex-wrap gap-2">
                  {storeTags.map((tag) => {
                    const isSelected = selectedTags.find((t) => t.id === tag.id);
                    return (
                      <button
                        key={tag.id}
                        onClick={() => toggleTag(tag)}
                        className={cn(
                          "px-3 py-1 rounded-full text-sm transition-colors",
                          isSelected
                            ? "text-white"
                            : "bg-muted text-muted-foreground hover:bg-muted/80"
                        )}
                        style={
                          isSelected
                            ? { backgroundColor: tag.color }
                            : undefined
                        }
                      >
                        {tag.name}
                      </button>
                    );
                  })}
                  <button
                    onClick={() => setIsAddingTag(true)}
                    className={cn(
                      "px-3 py-1 rounded-full text-sm transition-colors border border-dashed border-muted-foreground/30 text-muted-foreground hover:border-muted-foreground/50"
                    )}
                  >
                    <Plus className="w-3 h-3 inline mr-1" />
                    添加
                  </button>
                </div>
                {isAddingTag && (
                  <div className="mt-2 flex gap-2">
                    <Input
                      value={newTagName}
                      onChange={(e) => setNewTagName(e.target.value)}
                      placeholder="输入标签名称..."
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          e.preventDefault();
                          handleAddTag();
                        }
                      }}
                      className="flex-1"
                      autoFocus
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={handleAddTag}
                      disabled={!newTagName.trim()}
                    >
                      <Plus className="w-4 h-4 text-current" />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => {
                        setIsAddingTag(false);
                        setNewTagName("");
                      }}
                    >
                      <Trash2 className="w-4 h-4 text-current" />
                    </Button>
                  </div>
                )}
              </>
            </div>
          ) : (
            <div className="bg-muted/30 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-2">
                <Tag className="w-4 h-4 text-muted-foreground" />
                <span className="text-xs text-muted-foreground">标签</span>
              </div>
              <div className="flex flex-wrap gap-2">
                {selectedTags.length === 0 ? (
                  <span className="text-sm text-muted-foreground">暂无标签</span>
                ) : (
                  selectedTags.map((tag) => (
                    <span
                      key={tag.id}
                      className="px-3 py-1 rounded-full text-xs text-white"
                      style={{ backgroundColor: tag.color }}
                    >
                      {tag.name}
                    </span>
                  ))
                )}
              </div>
            </div>
          )}

          {/* 子任务 */}
          <div className={cn(!editing && "bg-muted/30 rounded-lg p-3")}>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium flex items-center gap-1.5">
                <ListChecks className="w-4 h-4 text-muted-foreground" />
                子任务
              </label>
              {aiEnabled && aiConfigured && (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleAIBreakdown}
                  disabled={aiBreakdownLoading}
                  className="h-7 gap-1 text-purple-600 hover:text-purple-700 hover:bg-purple-50 dark:hover:bg-purple-950 text-xs"
                >
                  {aiBreakdownLoading ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : (
                    <Sparkles className="w-3.5 h-3.5" />
                  )}
                  AI 拆解
                </Button>
              )}
            </div>
            {aiBreakdownError && (
              <p className="text-xs text-red-500 mb-2">{aiBreakdownError}</p>
            )}
            {!editing && subTasks.length > 0 && (
              <p className="text-xs text-muted-foreground mb-2">
                {subTasks.filter(st => st.completed).length}/{subTasks.length} 已完成
              </p>
            )}
            <div className="space-y-2">
              {subTasks.length === 0 && (
                <p className="text-sm text-muted-foreground py-1">暂无子任务</p>
              )}
              {subTasks.map((subTask) => (
                <div
                  key={subTask.id}
                  className="flex items-center gap-2 p-2 rounded-lg bg-muted/50"
                >
                  {editing ? (
                    <>
                      <Checkbox
                        checked={subTask.completed}
                        onChange={() => toggleSubTask(subTask.id)}
                      />
                      <span
                        className={cn(
                          "flex-1 text-sm",
                          subTask.completed && "line-through text-muted-foreground"
                        )}
                      >
                        {subTask.title}
                      </span>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => deleteSubTask(subTask.id)}
                        className="h-8 w-8"
                      >
                        <Trash2 className="w-4 h-4 text-muted-foreground" />
                      </Button>
                    </>
                  ) : (
                    <>
                      {subTask.completed ? (
                        <CheckCircle2 className="w-4 h-4 text-green-500 flex-shrink-0" />
                      ) : (
                        <Circle className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                      )}
                      <span
                        className={cn(
                          "flex-1 text-sm",
                          subTask.completed && "line-through text-muted-foreground"
                        )}
                      >
                        {subTask.title}
                      </span>
                    </>
                  )}
                </div>
              ))}
              {editing && (
                <div className="flex gap-2">
                  <Input
                    value={newSubTaskTitle}
                    onChange={(e) => setNewSubTaskTitle(e.target.value)}
                    placeholder="添加子任务..."
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        e.preventDefault();
                        addSubTask();
                      }
                    }}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={addSubTask}
                    disabled={!newSubTaskTitle.trim()}
                  >
                    <Plus className="w-4 h-4 text-current" />
                  </Button>
                </div>
              )}
            </div>
          </div>

          {/* 提醒 */}
          <div>
            <label className="text-sm font-medium mb-2 flex items-center gap-2">
              <Bell className="w-4 h-4 text-current" />
              提醒
            </label>
            {editing ? (
              <div className="space-y-3">
                {reminders.map((reminder) => (
                  <div
                    key={reminder.id}
                    className="flex items-center gap-2"
                  >
                    <DatePicker
                      value={new Date(reminder.date)}
                      onChange={(date) =>
                        updateReminder(reminder.id, {
                          date: date || new Date(),
                        })
                      }
                      placeholder="选择提醒时间"
                      className="flex-1"
                      showSeconds={true}
                    />
                    <Select
                      value={reminder.repeat}
                      onChange={(e) =>
                        updateReminder(reminder.id, {
                          repeat: e.target.value as Reminder["repeat"],
                        })
                      }
                      className="w-32"
                    >
                      <option value="none">不重复</option>
                      <option value="daily">每天</option>
                      <option value="weekly">每周</option>
                      <option value="monthly">每月</option>
                    </Select>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => deleteReminder(reminder.id)}
                      className="h-10 w-10 shrink-0"
                    >
                      <Trash2 className="w-4 h-4 text-muted-foreground" />
                    </Button>
                  </div>
                ))}
                <Button
                  type="button"
                  variant="outline"
                  className="w-full rounded-xl"
                  onClick={addReminder}
                >
                  <Plus className="w-4 h-4 mr-2 text-current" />
                  添加提醒
                </Button>
              </div>
            ) : (
              <div className="bg-muted/30 rounded-lg p-3 space-y-2">
                {reminders.length === 0 ? (
                  <p className="text-sm text-muted-foreground">暂无提醒</p>
                ) : (
                  reminders.map((reminder) => (
                    <div
                      key={reminder.id}
                      className="flex items-center gap-2 text-sm"
                    >
                      <span className="flex-1">{formatDate(reminder.date)}</span>
                      <span className="px-2 py-0.5 rounded text-xs bg-muted text-muted-foreground">
                        {repeatLabel(reminder.repeat || "none")}
                      </span>
                    </div>
                  ))
                )}
              </div>
            )}
          </div>
        </div>

        <DialogFooter className="sm:space-x-3 gap-3">
          {editing ? (
            <>
              <Button variant="outline" onClick={handleCancelEdit}>
                取消
              </Button>
              <Button onClick={handleSave}>保存</Button>
            </>
          ) : (
            <>
              <Button
                variant="outline"
                onClick={() => toggleTaskComplete(task.id)}
              >
                {task.completed ? "取消完成" : "完成"}
              </Button>
              <Button onClick={() => setEditing(true)}>编辑</Button>
            </>
          )}
        </DialogFooter>
      </DialogContent>

      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onOpenChange={setIsDeleteConfirmOpen}
        title={
          deleteTarget?.type === "task"
            ? "删除任务"
            : deleteTarget?.type === "subtask"
            ? "删除子任务"
            : "删除提醒"
        }
        description={
          deleteTarget?.type === "task"
            ? "确定要永久删除这个任务吗？此操作不可撤销。"
            : deleteTarget?.type === "subtask"
            ? "确定要删除这个子任务吗？"
            : "确定要删除这个提醒吗？"
        }
        confirmLabel="删除"
        variant="danger"
        onConfirm={handleConfirmDelete}
      />
    </Dialog>
  );
}
