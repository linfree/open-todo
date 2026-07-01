import { useState, useEffect } from "react";
import { Calendar, Sparkles, Loader2 } from "lucide-react";
import { useTodoStore } from "../store/todoStore";
import { Priority, TaskStatus } from "../types";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { cn } from "../lib/utils";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "./ui/dialog";
import { useAIStatus } from "../lib/ai";

interface AddTaskDialogProps {
  isOpen: boolean;
  onClose: () => void;
  dueDate?: Date;
}

export function AddTaskDialog({ isOpen, onClose, dueDate }: AddTaskDialogProps) {
  const [title, setTitle] = useState("");
  const [useAI, setUseAI] = useState(false);
  const [aiLoading, setAiLoading] = useState(false);
  const { addTask, currentListId, currentCategoryId } = useTodoStore();
  const { configured } = useAIStatus();

  useEffect(() => {
    if (isOpen) {
      setUseAI(configured);
    }
  }, [isOpen, configured]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;

    if (useAI && configured) {
      setAiLoading(true);
      try {
        const res = await fetch("/api/v1/ai/parse-task", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ input: title.trim() }),
        });
        if (res.ok) {
          const parsed = await res.json();
          addTask({
            title: parsed.title || title.trim(),
            description: parsed.description || undefined,
            completed: false,
            priority: (parsed.priority as Priority) || Priority.NONE,
            status: TaskStatus.TODO,
            listId: currentListId || "all",
            categoryId: currentCategoryId || undefined,
            tags: [],
            subTasks: [],
            reminders: parsed.reminder
              ? [{ id: crypto.randomUUID(), date: new Date(parsed.reminder), repeat: "none" as const, enabled: true }]
              : [],
            order: 0,
            dueDate: parsed.dueDate ? new Date(parsed.dueDate) : dueDate,
          });
        } else {
          // AI 失败，正常创建
          addTask({
            title: title.trim(),
            completed: false,
            priority: Priority.NONE,
            status: TaskStatus.TODO,
            listId: currentListId || "all",
            categoryId: currentCategoryId || undefined,
            tags: [],
            subTasks: [],
            reminders: [],
            order: 0,
            dueDate: dueDate,
          });
        }
      } catch {
        addTask({
          title: title.trim(),
          completed: false,
          priority: Priority.NONE,
          status: TaskStatus.TODO,
          listId: currentListId || "all",
          categoryId: currentCategoryId || undefined,
          tags: [],
          subTasks: [],
          reminders: [],
          order: 0,
          dueDate: dueDate,
        });
      } finally {
        setAiLoading(false);
      }
    } else {
      addTask({
        title: title.trim(),
        completed: false,
        priority: Priority.NONE,
        status: TaskStatus.TODO,
        listId: currentListId || "all",
        categoryId: currentCategoryId || undefined,
        tags: [],
        subTasks: [],
        reminders: [],
        order: 0,
        dueDate: dueDate,
      });
    }
    setTitle("");
    onClose();
  }

  return (
    <Dialog open={isOpen} onOpenChange={() => onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>添加任务</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="py-2 space-y-3">
            {configured && (
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setUseAI(!useAI)}
                  className={cn(
                    "flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium transition-all",
                    useAI
                      ? "bg-purple-100 text-purple-700 border border-purple-300 dark:bg-purple-900/30 dark:text-purple-300 dark:border-purple-700"
                      : "bg-muted text-muted-foreground border border-border hover:bg-muted/80"
                  )}
                >
                  <Sparkles className="w-3 h-3" />
                  AI 创建
                </button>
                {aiLoading && <Loader2 className="w-3.5 h-3.5 animate-spin text-purple-500" />}
              </div>
            )}
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={aiLoading ? "AI 正在解析..." : useAI ? "用自然语言描述任务，如：明天下午3点提醒我开会" : "准备做什么？"}
              className="text-base"
              autoFocus
              disabled={aiLoading}
            />
            {dueDate && (
              <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                <Calendar className="w-4 h-4" />
                <span>截止日期：{dueDate.toLocaleDateString("zh-CN")}</span>
              </div>
            )}
          </div>
          <DialogFooter className="mt-2">
            <Button type="button" variant="ghost" onClick={onClose} className="cursor-pointer">
              取消
            </Button>
            <Button type="submit" disabled={aiLoading}>
              {aiLoading ? "解析中..." : "添加"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
