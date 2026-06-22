import dayjs from "dayjs";

export const PRIORITY_OPTIONS = [
  { value: "high", label: "高" },
  { value: "normal", label: "中" },
  { value: "low", label: "低" },
];

export const PRIORITY_COLOR = {
  high: "#f04438",
  normal: "#2d7ff9",
  low: "#94a3b8",
};

export const STATUS_COLOR = {
  todo: "default",
  doing: "processing",
  done: "success",
  cancelled: "error",
  active: "processing",
  archived: "default",
};

export const STATUS_OPTIONS = [
  { value: "todo", label: "未开始" },
  { value: "doing", label: "进行中" },
  { value: "done", label: "已完成" },
  { value: "cancelled", label: "取消" },
];

export const STAGE_STATUS_OPTIONS = [
  { value: "active", label: "进行中" },
  { value: "done", label: "已完成" },
  { value: "archived", label: "归档" },
];

export const EXECUTE_MODE_OPTIONS = [
  { value: "weekly_todo", label: "本周待办" },
  { value: "scheduled_day", label: "指定某天执行" },
  { value: "split_to_days", label: "拆解为多个日任务" },
  { value: "ddl_only", label: "只设置 DDL" },
];

export const PROGRESS_MODE_OPTIONS = [
  { value: "weekly", label: "按周任务完成计算" },
  { value: "manual", label: "手动维护" },
];

export function toFormValues(type, item, fallback) {
  if (!item) {
    if (type === "daily") {
      return {
        priority: "normal",
        status: "todo",
        date: fallback.defaultDate ? dateObject(fallback.today) : undefined,
      };
    }
    if (type === "weekly") {
      return {
        priority: "normal",
        status: "todo",
        execute_mode: "weekly_todo",
        week_range: [dateObject(fallback.weekStart), dateObject(fallback.weekEnd)],
      };
    }
    return {
      status: "active",
      progress_mode: "weekly",
      stage_range: [dateObject(fallback.today), null],
    };
  }

  return Object.fromEntries(Object.entries({
    ...item,
    date: dateObject(getTaskDate(item)),
    deadline: dateObject(getDeadlineDate(item)),
    week_range: item.week_start ? [dateObject(item.week_start), dateObject(item.week_end) || null] : undefined,
    stage_range: item.start_date ? [dateObject(item.start_date), dateObject(item.end_date) || null] : undefined,
    target_text: getTargetText(item),
  }).filter(([, value]) => value !== undefined));
}

export function fromFormValues(type, values) {
  const payload = { ...values };
  const execution = { mode: values.execute_mode, date: formatMaybeDate(values.execute_date), dates: [] };

  if (type === "daily") {
    payload.date = formatMaybeDate(values.date);
    payload.deadline = formatMaybeDate(values.deadline);
    delete payload.week_range;
    delete payload.stage_range;
    delete payload.execute_date;
    delete payload.split_range;
  }

  if (type === "weekly") {
    payload.week_start = formatMaybeDate(values.week_range?.[0]);
    payload.week_end = formatMaybeDate(values.week_range?.[1]);
    payload.deadline = formatMaybeDate(values.deadline);
    payload.execute_mode = values.execute_mode || "weekly_todo";
    execution.mode = payload.execute_mode;
    execution.date = formatMaybeDate(values.execute_date);
    execution.dates = datesInRange(values.split_range);
    delete payload.week_range;
    delete payload.execute_date;
    delete payload.split_range;
    delete payload.stage_range;
  }

  if (type === "stage") {
    payload.start_date = formatMaybeDate(values.stage_range?.[0]);
    payload.end_date = formatMaybeDate(values.stage_range?.[1]);
    payload.progress_mode = values.progress_mode || "weekly";
    delete payload.stage_range;
    delete payload.week_range;
    delete payload.execute_date;
    delete payload.split_range;
  }

  return { payload, execution };
}

export function priorityLabel(priority) {
  return PRIORITY_OPTIONS.find((option) => option.value === priority)?.label || "中";
}

export function statusLabel(status) {
  return STATUS_OPTIONS.find((option) => option.value === status)?.label || status || "-";
}

export function stageStatusLabel(status) {
  return STAGE_STATUS_OPTIONS.find((option) => option.value === status)?.label || status || "进行中";
}

export function executeModeLabel(mode) {
  return EXECUTE_MODE_OPTIONS.find((option) => option.value === mode)?.label || "本周待办";
}

export function progressModeLabel(mode) {
  return PROGRESS_MODE_OPTIONS.find((option) => option.value === mode)?.label || "按周任务";
}

export function stageTitleById(stageGoals, id) {
  return stageGoals.find((goal) => goal.id === id)?.title || "未命名阶段";
}

export function getTaskDate(task) {
  return task?.date || task?.plan_date || "";
}

export function getDeadlineDate(task) {
  return task?.deadline || task?.due_date || "";
}

export function getTargetText(goal) {
  return goal?.target_text || goal?.target_metric || "";
}

export function isTaskDone(task) {
  return task?.done || task?.status === "done";
}

export function taskOriginLabel(task) {
  if (task?.origin === "weekly_materialized") return "周任务生成";
  if (task?.weekly_task_id) return "周任务关联";
  if (task?.origin === "manual") return "手动";
  return "";
}

export function taskOriginColor(task) {
  if (task?.origin === "weekly_materialized") return "processing";
  if (task?.weekly_task_id) return "geekblue";
  if (task?.origin === "manual") return "default";
  return "default";
}

export function isSameDate(value, dateKey) {
  return formatDate(value) === dateKey;
}

export function datesInRange(range) {
  if (!range?.[0] || !range?.[1]) return [];
  const dates = [];
  let cursor = range[0];
  const end = range[1];
  while (cursor.isBefore(end) || cursor.isSame(end, "day")) {
    dates.push(cursor.format("YYYY-MM-DD"));
    cursor = cursor.add(1, "day");
    if (dates.length >= 14) break;
  }
  return dates;
}

export function formatMinutes(minutes) {
  if (!minutes) return "0m";
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  if (hours === 0) return `${rest}m`;
  if (rest === 0) return `${hours}h`;
  return `${hours}h ${rest}m`;
}

export function formatTime(date) {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

export function formatSeconds(date) {
  return `:${String(date.getSeconds()).padStart(2, "0")}`;
}

export function formatLongDate(date) {
  return `${date.getFullYear()} 年 ${date.getMonth() + 1} 月 ${date.getDate()} 日`;
}

export function weekdayLabel(date) {
  return ["周日", "周一", "周二", "周三", "周四", "周五", "周六"][date.getDay()];
}

export function formatDateKey(date) {
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export function startOfWeek(date) {
  const result = new Date(date);
  const day = result.getDay() || 7;
  result.setDate(result.getDate() - day + 1);
  return result;
}

export function endOfWeek(date) {
  const result = startOfWeek(date);
  result.setDate(result.getDate() + 6);
  return result;
}

export function formatDate(value) {
  if (!value) return "";
  return formatDateKey(new Date(value));
}

export function dateObject(value) {
  if (!value) return undefined;
  return dayjs(value);
}

export function formatMaybeDate(value) {
  if (!value) return "";
  return value.format ? value.format("YYYY-MM-DD") : formatDateKey(new Date(value));
}

export function formatLogRange(log) {
  if (!log) return "";
  const start = log.start_time ? new Date(log.start_time) : null;
  const end = log.end_time ? new Date(log.end_time) : null;
  if (start && end) return `${formatTime(start)} - ${formatTime(end)}`;
  if (log.duration_min) return `${log.duration_min} 分钟`;
  return "";
}
