import { useState, FormEvent } from "react";
import { Sparkles, Loader2 } from "lucide-react";
import { useTodoStore } from "../store/todoStore";
import { Priority, TaskStatus } from "../types";
import { useAIStatus } from "../lib/ai";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { cn } from "../lib/utils";

interface ParsedTaskResponse {
  title: string;
  description: string;
  priority: string;
  dueDate: string | null;
  reminder: string | null;
}

function mapPriority(p: string): Priority {
  switch (p) {
    case "low": return Priority.LOW;
    case "medium": return Priority.MEDIUM;
    case "high": return Priority.HIGH;
    default: return Priority.NONE;
  }
}

export function AITaskInput() {
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { enabled, configured, loading: statusLoading } = useAIStatus();
  const { addTask, currentListId, currentCategoryId } = useTodoStore();

  // Don't show if AI is not available
  if (statusLoading) return null;
  if (!enabled || !configured) return null;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed) return;

    setLoading(true);
    setError(null);

    try {
      const res = await fetch("/api/v1/ai/parse-task", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ input: trimmed }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: `HTTP ${res.status}` }));
        throw new Error(data.error || `请求失败 (${res.status})`);
      }

      const parsed: ParsedTaskResponse = await res.json();

      addTask({
        title: parsed.title || trimmed,
        description: parsed.description || undefined,
        completed: false,
        priority: mapPriority(parsed.priority),
        status: TaskStatus.TODO,
        listId: currentListId || "all",
        categoryId: currentCategoryId || undefined,
        tags: [],
        subTasks: [],
        reminders: parsed.reminder
          ? [{ id: crypto.randomUUID(), date: new Date(parsed.reminder), repeat: "none", enabled: true }]
          : [],
        order: 0,
        dueDate: parsed.dueDate ? new Date(parsed.dueDate) : undefined,
      });

      setInput("");
    } catch (err: any) {
      const msg = typeof err === "string" ? err : (err.message || "未知错误");
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="w-full max-w-4xl 2xl:max-w-6xl">
      <form onSubmit={handleSubmit} className="flex items-center gap-2">
        <div className="relative flex-1">
          <Sparkles className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-purple-500 pointer-events-none" />
          <Input
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              if (error) setError(null);
            }}
            placeholder="用一句话创建任务，如：明天下午3点提醒我开会"
            className="pl-10 h-10 text-sm"
            disabled={loading}
          />
        </div>
        <Button
          type="submit"
          disabled={loading || !input.trim()}
          size="sm"
          className={cn(
            "gap-1.5 shrink-0",
            "bg-purple-600 hover:bg-purple-700 text-white"
          )}
        >
          {loading ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <Sparkles className="w-4 h-4" />
          )}
          AI 创建
        </Button>
      </form>
      {error && (
        <p className="text-xs text-red-500 mt-1.5 ml-1">{error}</p>
      )}
    </div>
  );
}
