import { useEffect, useMemo, useState } from "react";
import { Form, Tabs, message } from "antd";
import { CalendarOutlined, FieldTimeOutlined, FlagOutlined } from "@ant-design/icons";
import { Page, PageHeader } from "../components/ui";
import { getTodayPomodoroStats } from "../api/pomodoro";
import { getLogStats, getLogs } from "../api/logs";
import {
  createPlanningDailyTask,
  createStageGoal,
  createWeeklyTask,
  deletePlanningDailyTask,
  deleteStageGoal,
  deleteWeeklyTask,
  listPlanningDailyTasks,
  listStageGoals,
  listWeeklyTasks,
  materializeWeeklyTask,
  togglePlanningDailyTask,
  updatePlanningDailyTask,
  updateStageGoal,
  updateWeeklyTask,
} from "../api/planning";
import { PlanningModal } from "./dashboard/DashboardComponents";
import { StageTab, TodayTab, WeekTab } from "./dashboard/DashboardSections";
import {
  endOfWeek,
  formatDateKey,
  fromFormValues,
  getDeadlineDate,
  isSameDate,
  isTaskDone,
  startOfWeek,
  toFormValues,
} from "./dashboard/dashboardUtils";

function Dashboard() {
  const [now, setNow] = useState(() => new Date());
  const [dailyTasks, setDailyTasks] = useState([]);
  const [inboxTasks, setInboxTasks] = useState([]);
  const [deadlineTasks, setDeadlineTasks] = useState([]);
  const [weeklyTasks, setWeeklyTasks] = useState([]);
  const [stageGoals, setStageGoals] = useState([]);
  const [pomodoroStats, setPomodoroStats] = useState(null);
  const [logStats, setLogStats] = useState(null);
  const [recentLogs, setRecentLogs] = useState([]);
  const [dailyTitle, setDailyTitle] = useState("");
  const [weeklyTitle, setWeeklyTitle] = useState("");
  const [stageTitle, setStageTitle] = useState("");
  const [detailModal, setDetailModal] = useState({ open: false, type: "daily", item: null });
  const [form] = Form.useForm();

  const today = useMemo(() => formatDateKey(now), [now]);
  const weekStart = useMemo(() => formatDateKey(startOfWeek(now)), [now]);
  const weekEnd = useMemo(() => formatDateKey(endOfWeek(now)), [now]);

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  const loadPlanning = () => {
    Promise.all([
      listPlanningDailyTasks({ date: today }).catch(() => []),
      listPlanningDailyTasks({ unscheduled: true }).catch(() => []),
      listPlanningDailyTasks({ deadline: today }).catch(() => []),
      listWeeklyTasks({ week_start: weekStart }).catch(() => []),
      listStageGoals().catch(() => []),
    ]).then(([daily, inbox, deadline, weekly, stages]) => {
      setDailyTasks(Array.isArray(daily) ? daily : []);
      setInboxTasks(Array.isArray(inbox) ? inbox : []);
      setDeadlineTasks(Array.isArray(deadline) ? deadline : []);
      setWeeklyTasks(Array.isArray(weekly) ? weekly : []);
      setStageGoals(Array.isArray(stages) ? stages : []);
    });
  };

  const loadStats = () => {
    Promise.all([
      getTodayPomodoroStats().catch(() => null),
      getLogStats({ date: today, scope: "day" }).catch(() => null),
      getLogs({ date: today, scope: "day" }).catch(() => []),
    ]).then(([pomo, log, logs]) => {
      setPomodoroStats(pomo);
      setLogStats(log);
      setRecentLogs(Array.isArray(logs) ? logs.slice(0, 4) : []);
    });
  };

  useEffect(() => {
    loadPlanning();
    loadStats();
    const onPomodoro = () => loadStats();
    window.addEventListener("pomodoro:updated", onPomodoro);
    return () => window.removeEventListener("pomodoro:updated", onPomodoro);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [today, weekStart]);

  const doneCount = dailyTasks.filter((task) => isTaskDone(task)).length;
  const totalCount = dailyTasks.length;
  const pendingCount = totalCount - doneCount;
  const completionPercent = totalCount === 0 ? 0 : Math.round((doneCount / totalCount) * 100);
  const focusMinutes = logStats?.total_minutes ?? pomodoroStats?.focus_minutes ?? 0;
  const focusCount = pomodoroStats?.focus_count ?? 0;
  const mainSubject = logStats?.main_subject || "-";
  const stageLinkedWeeklyTasks = weeklyTasks.filter((task) => task.stage_goal_id);
  const independentWeeklyTasks = weeklyTasks.filter((task) => !task.stage_goal_id && task.execute_mode !== "ddl_only");
  const weeklyInboxTasks = weeklyTasks.filter((task) => task.execute_mode === "ddl_only");
  const weeklyDoneCount = weeklyTasks.filter((task) => task.status === "done" || task.progress_percent >= 100).length;
  const weeklyProgressPercent = weeklyTasks.length
    ? Math.round(weeklyTasks.reduce((sum, task) => sum + (task.progress_percent || 0), 0) / weeklyTasks.length)
    : 0;
  const activeStageCount = stageGoals.filter((goal) => goal.status !== "done" && goal.status !== "archived").length;
  const completedStageCount = stageGoals.filter((goal) => goal.status === "done").length;
  const stageProgressPercent = stageGoals.length
    ? Math.round(stageGoals.reduce((sum, goal) => sum + (goal.progress_percent || 0), 0) / stageGoals.length)
    : 0;
  const todayDeadlineItems = [
    ...deadlineTasks
      .filter((task) => !isTaskDone(task))
      .map((task) => ({ type: "daily", item: task })),
    ...weeklyTasks
      .filter((task) => !isTaskDone(task) && isSameDate(getDeadlineDate(task), today))
      .map((task) => ({ type: "weekly", item: task })),
  ];

  const addDailyFast = async () => {
    const title = dailyTitle.trim();
    if (!title) return;
    try {
      await createPlanningDailyTask({ title, date: today, priority: "normal", status: "todo" });
      setDailyTitle("");
      loadPlanning();
      message.success("已添加今日任务");
    } catch {
      message.error("添加今日任务失败");
    }
  };

  const addWeeklyFast = async () => {
    const title = weeklyTitle.trim();
    if (!title) return;
    try {
      await createWeeklyTask({
        title,
        week_start: weekStart,
        week_end: weekEnd,
        priority: "normal",
        status: "todo",
        execute_mode: "weekly_todo",
      });
      setWeeklyTitle("");
      loadPlanning();
      message.success("已添加周任务");
    } catch {
      message.error("添加周任务失败");
    }
  };

  const addStageFast = async () => {
    const title = stageTitle.trim();
    if (!title) return;
    try {
      await createStageGoal({ title, start_date: today, status: "active", progress_mode: "weekly" });
      setStageTitle("");
      loadPlanning();
      message.success("已添加阶段目标");
    } catch {
      message.error("添加阶段目标失败");
    }
  };

  const toggleDaily = async (task) => {
    const nextDone = !isTaskDone(task);
    setDailyTasks((prev) => prev.map((item) => (item.id === task.id ? { ...item, done: nextDone, status: nextDone ? "done" : "todo" } : item)));
    try {
      await togglePlanningDailyTask(task.id);
      loadPlanning();
    } catch {
      setDailyTasks((prev) => prev.map((item) => (item.id === task.id ? task : item)));
      message.error("更新任务状态失败");
    }
  };

  const openDetail = (type, item = null, options = {}) => {
    setDetailModal({ open: true, type, item });
    form.resetFields();
    form.setFieldsValue(toFormValues(type, item, {
      today,
      weekStart,
      weekEnd,
      defaultDate: options.defaultDate !== false,
    }));
  };

  const closeDetail = () => setDetailModal({ open: false, type: "daily", item: null });

  const saveDetail = async () => {
    const values = await form.validateFields();
    const { payload, execution } = fromFormValues(detailModal.type, values);
    try {
      let savedItem = null;
      if (detailModal.type === "daily") {
        savedItem = detailModal.item
          ? await updatePlanningDailyTask(detailModal.item.id, payload)
          : await createPlanningDailyTask(payload);
      }
      if (detailModal.type === "weekly") {
        savedItem = detailModal.item
          ? await updateWeeklyTask(detailModal.item.id, payload)
          : await createWeeklyTask(payload);
        await applyWeeklyExecution(savedItem, execution);
      }
      if (detailModal.type === "stage") {
        savedItem = detailModal.item
          ? await updateStageGoal(detailModal.item.id, payload)
          : await createStageGoal(payload);
      }
      closeDetail();
      loadPlanning();
      message.success("已保存");
    } catch {
      message.error("保存失败");
    }
  };

  const arrangeWeeklyToToday = async (task) => {
    try {
      await materializeWeeklyTask(task.id, { mode: "scheduled_day", date: today, replace: false });
      loadPlanning();
      message.success("已安排到今天");
    } catch {
      message.error("安排失败");
    }
  };

  const applyWeeklyExecution = async (weeklyTask, execution) => {
    if (!weeklyTask || !execution) return;
    const mode = execution.mode || "weekly_todo";
    if (execution.mode === "scheduled_day" && execution.date) {
      await materializeWeeklyTask(weeklyTask.id, { mode, date: execution.date, replace: true });
      return;
    }
    if (execution.mode === "split_to_days" && execution.dates.length) {
      await materializeWeeklyTask(weeklyTask.id, { mode, dates: execution.dates, replace: true });
      return;
    }
    if (mode === "weekly_todo" || mode === "ddl_only") {
      await materializeWeeklyTask(weeklyTask.id, { mode, replace: true });
    }
  };

  const removeItem = async (type, item) => {
    try {
      if (type === "daily") await deletePlanningDailyTask(item.id);
      if (type === "weekly") await deleteWeeklyTask(item.id);
      if (type === "stage") await deleteStageGoal(item.id);
      loadPlanning();
    } catch {
      message.error("删除失败");
    }
  };

  return (
    <Page className="dashboard-page">
      <PageHeader
        eyebrow="Overview"
        title="总览"
        description="把长期目标拆到每周，再落到今天真正要完成的事。"
      />
      <Tabs
        className="dashboard-tabs app-section-tabs"
        items={[
          {
            key: "today",
            label: <span className="dashboard-tab-label"><CalendarOutlined />日计划</span>,
            children: (
              <TodayTab
                now={now}
                dailyTasks={dailyTasks}
                dailyTitle={dailyTitle}
                doneCount={doneCount}
                totalCount={totalCount}
                pendingCount={pendingCount}
                completionPercent={completionPercent}
                focusMinutes={focusMinutes}
                focusCount={focusCount}
                mainSubject={mainSubject}
                recentLogs={recentLogs}
                todayDeadlineItems={todayDeadlineItems}
                onDailyTitleChange={setDailyTitle}
                onAddDaily={addDailyFast}
                onToggleDaily={toggleDaily}
                onEditDaily={(task) => openDetail("daily", task)}
                onDeleteDaily={(task) => removeItem("daily", task)}
                onEditWeekly={(task) => openDetail("weekly", task)}
                onArrangeWeekly={arrangeWeeklyToToday}
                onCreateDaily={() => openDetail("daily")}
              />
            ),
          },
          {
            key: "week",
            label: <span className="dashboard-tab-label"><FieldTimeOutlined />周计划</span>,
            children: (
              <WeekTab
                weekStart={weekStart}
                weekEnd={weekEnd}
                weeklyTitle={weeklyTitle}
                weeklyTasks={weeklyTasks}
                stageLinkedWeeklyTasks={stageLinkedWeeklyTasks}
                independentWeeklyTasks={independentWeeklyTasks}
                weeklyInboxTasks={weeklyInboxTasks}
                inboxTasks={inboxTasks}
                stageGoals={stageGoals}
                weeklyProgressPercent={weeklyProgressPercent}
                weeklyDoneCount={weeklyDoneCount}
                onWeeklyTitleChange={setWeeklyTitle}
                onAddWeekly={addWeeklyFast}
                onCreateWeekly={() => openDetail("weekly")}
                onEditWeekly={(task) => openDetail("weekly", task)}
                onDeleteWeekly={(task) => removeItem("weekly", task)}
                onArrangeWeekly={arrangeWeeklyToToday}
                onToggleDaily={toggleDaily}
                onEditDaily={(task) => openDetail("daily", task)}
                onDeleteDaily={(task) => removeItem("daily", task)}
                onCreateInboxDaily={() => openDetail("daily", null, { defaultDate: false })}
              />
            ),
          },
          {
            key: "stage",
            label: <span className="dashboard-tab-label"><FlagOutlined />长期规划</span>,
            children: (
              <StageTab
                stageTitle={stageTitle}
                stageGoals={stageGoals}
                activeStageCount={activeStageCount}
                completedStageCount={completedStageCount}
                stageProgressPercent={stageProgressPercent}
                onStageTitleChange={setStageTitle}
                onAddStage={addStageFast}
                onCreateStage={() => openDetail("stage")}
                onEditStage={(goal) => openDetail("stage", goal)}
                onDeleteStage={(goal) => removeItem("stage", goal)}
              />
            ),
          },
        ]}
      />

      <PlanningModal
        modal={detailModal}
        form={form}
        stageGoals={stageGoals}
        weeklyTasks={weeklyTasks}
        onCancel={closeDetail}
        onSave={saveDetail}
      />
    </Page>
  );
}

export default Dashboard;
