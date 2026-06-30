import { useState } from "react";
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
}

export function AddTaskDialog({ isOpen, onClose }: AddTaskDialogProps) {
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
    });
    setTitle("");
    onClose();
  }

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>添加任务</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="py-2">
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="准备做什么？"
              className="text-base"
              autoFocus
            />
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
