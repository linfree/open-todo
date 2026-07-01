import { useState, useMemo } from "react";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { useTodoStore } from "../store/todoStore";
import { Task } from "../types";
import { Button } from "./ui/button";
import { cn } from "../lib/utils";

const WEEKDAYS = ["日", "一", "二", "三", "四", "五", "六"] as const;

interface CalendarViewProps {
  onTaskClick: (task: Task) => void;
  onDateClick?: (date: Date) => void;
}

export function CalendarView({ onTaskClick, onDateClick }: CalendarViewProps) {
  const { getFilteredTasks } = useTodoStore();
  const tasks = getFilteredTasks();
  const [currentDate, setCurrentDate] = useState(new Date());

  const { daysInMonth, firstDayOfMonth, currentMonth, currentYear } = useMemo(() => {
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();
    const firstDay = new Date(year, month, 1).getDay();
    const days = new Date(year, month + 1, 0).getDate();
    return { daysInMonth: days, firstDayOfMonth: firstDay, currentMonth: month, currentYear: year };
  }, [currentDate]);

  const getTasksForDate = (day: number) => {
    return tasks.filter((task) => {
      if (!task.dueDate) return false;
      const taskDate = new Date(task.dueDate);
      return taskDate.getDate() === day && taskDate.getMonth() === currentMonth && taskDate.getFullYear() === currentYear;
    });
  };

  const prevMonth = () => setCurrentDate(new Date(currentYear, currentMonth - 1, 1));
  const nextMonth = () => setCurrentDate(new Date(currentYear, currentMonth + 1, 1));
  const goToToday = () => setCurrentDate(new Date());

  const calendarDays = [];
  for (let i = 0; i < firstDayOfMonth; i++) {
    calendarDays.push(<div key={`empty-${i}`} className="h-24 sm:h-28 bg-muted/20" />);
  }
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(currentYear, currentMonth, day);
    calendarDays.push(
      <CalendarDay
        key={day}
        day={day}
        date={date}
        tasks={getTasksForDate(day)}
        isToday={
          new Date().getDate() === day &&
          new Date().getMonth() === currentMonth &&
          new Date().getFullYear() === currentYear
        }
        onTaskClick={onTaskClick}
        onDateClick={onDateClick}
      />
    );
  }

  return (
    <div className="flex-1 overflow-y-auto scrollbar-hide p-4 sm:p-6 flex flex-col min-h-0">
      <div className="flex items-center justify-between mb-4 sm:mb-6">
        <div className="flex items-center gap-3 sm:gap-4">
          <h2 className="text-lg sm:text-xl font-semibold text-foreground">
            {currentYear}年{currentMonth + 1}月
          </h2>
          <Button variant="outline" size="sm" onClick={goToToday} className="cursor-pointer">今天</Button>
        </div>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" onClick={() => setCurrentDate(new Date(currentYear - 1, currentMonth, 1))} className="cursor-pointer" title="上一年">
            <ChevronsLeft className="w-4 h-4 text-current" />
          </Button>
          <Button variant="ghost" size="icon" onClick={prevMonth} className="cursor-pointer">
            <ChevronLeft className="w-4 h-4 text-current" />
          </Button>
          <select value={currentMonth} onChange={(e) => setCurrentDate(new Date(currentYear, parseInt(e.target.value), 1))} className="h-9 px-2 py-1 text-sm rounded-lg border border-border bg-muted/50 text-foreground cursor-pointer focus:outline-none focus:ring-2 focus:ring-ring appearance-none" aria-label="选择月份">
            {Array.from({ length: 12 }, (_, i) => (
              <option key={i} value={i}>{i + 1}月</option>
            ))}
          </select>
          <Button variant="ghost" size="icon" onClick={nextMonth} className="cursor-pointer">
            <ChevronRight className="w-4 h-4 text-current" />
          </Button>
          <Button variant="ghost" size="icon" onClick={() => setCurrentDate(new Date(currentYear + 1, currentMonth, 1))} className="cursor-pointer" title="下一年">
            <ChevronsRight className="w-4 h-4 text-current" />
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-7 gap-px bg-border/50 border border-border/50 rounded-t-xl overflow-hidden">
        {WEEKDAYS.map((day) => (
          <div key={day} className="bg-muted/40 p-2 text-center text-sm font-medium text-foreground">{day}</div>
        ))}
      </div>

      <div className="grid grid-cols-7 gap-px bg-border/50 border-x border-border/50 rounded-b-xl">
        {calendarDays}
      </div>

    </div>
  );
}

interface CalendarDayProps {
  day: number;
  date: Date;
  tasks: Task[];
  isToday: boolean;
  onTaskClick: (task: Task) => void;
  onDateClick?: (date: Date) => void;
}

function CalendarDay({ day, date, tasks, isToday, onTaskClick, onDateClick }: CalendarDayProps) {
  const maxTasksToShow = 3;
  const displayTasks = tasks.slice(0, maxTasksToShow);
  const remainingCount = tasks.length - maxTasksToShow;

  const priorityColors = {
    none: "bg-muted text-foreground",
    low: "bg-blue-500 text-white",
    medium: "bg-amber-500 text-white",
    high: "bg-red-500 text-white",
  };

  return (
    <div
      onClick={() => onDateClick?.(date)}
      className={cn(
        "h-24 sm:h-28 bg-card p-1.5 hover:bg-accent/30 transition-colors cursor-pointer",
        isToday && "bg-primary/5 ring-inset ring-2 ring-primary/20"
      )}
    >
      <div className={cn(
        "text-sm font-medium text-foreground mb-1.5 w-7 h-7 flex items-center justify-center rounded-lg mx-auto",
        isToday && "bg-primary text-primary-foreground shadow-sm"
      )}>
        {day}
      </div>
      <div className="space-y-1">
        {displayTasks.map((task) => (
          <div key={task.id} onClick={(e) => { e.stopPropagation(); onTaskClick(task); }}
            className={cn("text-xs px-1.5 py-0.5 rounded-md truncate shadow-sm", priorityColors[task.priority], task.completed && "opacity-50 line-through")}>
            {task.title}
          </div>
        ))}
        {remainingCount > 0 && (
          <div className="text-xs text-muted-foreground/70 px-1.5">+{remainingCount} 更多</div>
        )}
      </div>
    </div>
  );
}
