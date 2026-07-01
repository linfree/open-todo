import { useState } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Bell, Calendar, CheckCircle, GripVertical, RotateCcw, Trash2 } from "lucide-react";
import { useTodoStore } from "../store/todoStore";
import { Task } from "../types";
import { Badge } from "./ui/badge";
import { cn } from "../lib/utils";
import { PRIORITY_COLORS } from "../lib/icons";
import { ConfirmDialog } from "./ui/confirm-dialog";

interface SortableTaskItemProps {
  task: Task;
  onClick: () => void;
  onReminderClick?: () => void;
  isTrashView?: boolean;
}

export function SortableTaskItem({ task, onClick, onReminderClick, isTrashView }: SortableTaskItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: task.id });

  const { toggleTaskComplete, restoreTask, permanentlyDeleteTask } = useTodoStore();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // Touch swipe state
  const [touchStartX, setTouchStartX] = useState(0);
  const [swipeOffset, setSwipeOffset] = useState(0);
  const [isSwiping, setIsSwiping] = useState(false);

  const SWIPE_THRESHOLD = 80;

  function handleTouchStart(e: React.TouchEvent) {
    setTouchStartX(e.touches[0].clientX);
    setSwipeOffset(0);
    setIsSwiping(false);
  }

  function handleTouchMove(e: React.TouchEvent) {
    const diff = e.touches[0].clientX - touchStartX;
    if (Math.abs(diff) > 10) {
      setIsSwiping(true);
    }
    if (Math.abs(diff) > 5) {
      setSwipeOffset(diff);
    }
  }

  function handleTouchEnd() {
    if (isSwiping) {
      if (swipeOffset > SWIPE_THRESHOLD) {
        // Right swipe: mark complete
        toggleTaskComplete(task.id);
      } else if (swipeOffset < -SWIPE_THRESHOLD) {
        // Left swipe: delete
        permanentlyDeleteTask(task.id);
      }
    }
    setSwipeOffset(0);
    setTouchStartX(0);
    setIsSwiping(false);
  }

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  function handleRestore(e: React.MouseEvent) {
    e.stopPropagation();
    restoreTask(task.id);
  }

  function handlePermanentDelete(e: React.MouseEvent) {
    e.stopPropagation();
    setShowDeleteConfirm(true);
  }

  function getPriorityLabel() {
    switch (task.priority) {
      case "high":
        return "高";
      case "medium":
        return "中";
      case "low":
        return "低";
      default:
        return null;
    }
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes}>
      {isTrashView ? (
        /* Trash view: restore + permanent delete buttons */
        <div
          className={cn(
            "group flex items-start gap-2 sm:gap-3 p-3 sm:p-4 rounded-xl transition-all duration-200 cursor-pointer border border-transparent",
            isDragging ? "opacity-50" : "hover:bg-accent/50 hover:border-border/50"
          )}
          onClick={onClick}
        >
          <button
            onClick={handleRestore}
            className="mt-1 p-1 sm:p-1.5 rounded-lg hover:bg-green-500/10 hover:text-green-600 dark:hover:text-green-400 transition-all duration-200 flex-shrink-0 text-current"
            title="恢复任务"
          >
            <RotateCcw className="w-4 h-4" />
          </button>
          <button
            onClick={handlePermanentDelete}
            className="mt-1 p-1 sm:p-1.5 rounded-lg hover:bg-red-500/10 hover:text-red-600 dark:hover:text-red-400 transition-all duration-200 flex-shrink-0 text-current"
            title="永久删除"
          >
            <Trash2 className="w-4 h-4" />
          </button>

          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-foreground">
              {task.title}
            </p>
            {task.description && (
              <p className="text-xs text-muted-foreground mt-1 line-clamp-1 sm:line-clamp-2">
                {task.description}
              </p>
            )}
          </div>

          {task.priority !== "none" && (
            <Badge
              variant="outline"
              className={cn("text-xs font-medium flex-shrink-0", PRIORITY_COLORS[task.priority])}
            >
              {getPriorityLabel()}
            </Badge>
          )}
        </div>
      ) : (
        /* Active task: swipeable with drag handle and checkbox */
        <div
          className="relative overflow-hidden rounded-xl"
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
        >
          {/* Background: left green "完成", right red "删除" */}
          <div className="absolute inset-y-0 left-0 flex items-center justify-start pl-3">
            <div
              className={cn(
                "flex items-center gap-1 rounded-md px-3 py-1.5 text-white transition-opacity duration-150",
                "bg-green-500",
                swipeOffset > 0 ? "opacity-100" : "opacity-0"
              )}
            >
              <CheckCircle className="w-4 h-4" />
              <span className="text-sm font-medium">完成</span>
            </div>
          </div>
          <div className="absolute inset-y-0 right-0 flex items-center justify-end pr-3">
            <div
              className={cn(
                "flex items-center gap-1 rounded-md px-3 py-1.5 text-white transition-opacity duration-150",
                "bg-red-500",
                swipeOffset < 0 ? "opacity-100" : "opacity-0"
              )}
            >
              <Trash2 className="w-4 h-4" />
              <span className="text-sm font-medium">删除</span>
            </div>
          </div>

          {/* Sliding content */}
          <div
            className={cn(
              "group flex items-start gap-2 sm:gap-3 p-3 sm:p-4 rounded-xl transition-colors duration-200 cursor-pointer border border-transparent relative bg-card",
              isDragging ? "opacity-50" : "hover:bg-accent/50 hover:border-border/50"
            )}
            style={{
              transform: `translateX(${swipeOffset}px)`,
              transition: isSwiping ? "none" : "transform 0.2s ease-out",
              touchAction: "pan-y",
            }}
            onClick={onClick}
          >
            <button
              {...listeners}
              className="mt-1 cursor-grab active:cursor-grabbing flex-shrink-0 opacity-40 hover:opacity-70 transition-opacity text-current"
              aria-label="拖动排序"
            >
              <GripVertical className="w-4 h-4" />
            </button>
            <label
              onClick={(e) => e.stopPropagation()}
              className={cn(
                "mt-0.5 w-5 h-5 rounded-full border-2 flex-shrink-0 cursor-pointer flex items-center justify-center transition-all duration-200",
                task.completed
                  ? "bg-primary border-primary"
                  : "border-muted-foreground/30 hover:border-primary hover:scale-110 active:scale-95",
              )}
            >
              <input
                type="checkbox"
                checked={task.completed}
                onChange={(e) => {
                  e.stopPropagation();
                  toggleTaskComplete(task.id);
                }}
                className="sr-only"
                aria-checked={task.completed}
              />
              {task.completed && (
                <svg className="w-3 h-3 text-primary-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              )}
            </label>

            <div className="flex-1 min-w-0">
              <p
                className={cn(
                  "text-sm font-medium text-foreground transition-all duration-200",
                  task.completed && "text-muted-foreground line-through opacity-70"
                )}
              >
                {task.title}
              </p>
              {task.description && (
                <p className="text-xs text-muted-foreground mt-1 line-clamp-1 sm:line-clamp-2">
                  {task.description}
                </p>
              )}
              {(task.dueDate || task.tags.length > 0 || task.reminders.length > 0) && (
                <div className="flex items-center gap-2 mt-2 flex-wrap">
                  {task.dueDate && (
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Calendar className="w-3 h-3 text-current" />
                      <span>{new Date(task.dueDate).toLocaleDateString()}</span>
                    </div>
                  )}
                  {task.tags.map((tag) => (
                    <Badge key={tag.id} variant="outline" className="text-xs px-1.5 py-0 border-border/50">
                      {tag.name}
                    </Badge>
                  ))}
                  {task.reminders.length > 0 && onReminderClick && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onReminderClick();
                      }}
                      className="flex items-center gap-1 text-xs text-primary hover:text-primary/80 transition-colors"
                      title={`${task.reminders.length} 个提醒`}
                    >
                      <Bell className="w-3 h-3" />
                      <span>{task.reminders.length}</span>
                    </button>
                  )}
                </div>
              )}
            </div>

            {task.priority !== "none" && (
              <Badge
                variant="outline"
                className={cn("text-xs font-medium flex-shrink-0", PRIORITY_COLORS[task.priority])}
              >
                {getPriorityLabel()}
              </Badge>
            )}
          </div>
        </div>
      )}
      {showDeleteConfirm && (
        <ConfirmDialog
          open={showDeleteConfirm}
          onOpenChange={setShowDeleteConfirm}
          title="永久删除任务"
          description="确定要永久删除此任务吗？此操作无法撤销。"
          variant="danger"
          confirmLabel="删除"
          onConfirm={() => permanentlyDeleteTask(task.id)}
        />
      )}
    </div>
  );
}
