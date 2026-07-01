import { useState } from "react";
import { Calendar } from "lucide-react";
import { useTodoStore } from "../store/todoStore";
import { Priority, TaskStatus } from "../types";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "./ui/dialog";

interface AddTaskDialogProps {
  isOpen: boolean;
  onClose: () => void;
  dueDate?: Date;
}

export function AddTaskDialog({ isOpen, onClose, dueDate }: AddTaskDialogProps) {
  const [title, setTitle] = useState("");
  const { addTask, currentListId, currentCategoryId } = useTodoStore();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;

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
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="准备做什么？"
              className="text-base"
              autoFocus
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
            <Button type="submit">添加</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
