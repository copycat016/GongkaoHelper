import { useEffect, useMemo, useState } from "react";
import {
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Empty,
  Form,
  Input,
  List,
  Modal,
  Progress,
  Row,
  Select,
  Space,
  Tabs,
  Tag,
  Tooltip,
  message,
} from "antd";
import {
  CalendarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  FieldTimeOutlined,
  FlagOutlined,
  PlusOutlined,
  UnorderedListOutlined,
} from "@ant-design/icons";
import dayjs from "dayjs";
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
  togglePlanningDailyTask,
  updatePlanningDailyTask,
  updateStageGoal,
  updateWeeklyTask,
} from "../api/planning";

const PRIORITY_OPTIONS = [
  { value: "high", label: "高" },
  { value: "normal", label: "中" },
  { value: "low", label: "低" },
];

const PRIORITY_COLOR = {
  high: "#f04438",
  normal: "#2d7ff9",
  low: "#94a3b8",
};

const STATUS_COLOR = {
  todo: "default",
  doing: "processing",
  done: "success",
  cancelled: "error",
  active: "processing",
  archived: "default",
};

const STATUS_OPTIONS = [
  { value: "todo", label: "未开始" },
  { value: "doing", label: "进行中" },
  { value: "done", label: "已完成" },
  { value: "cancelled", label: "取消" },
];

const STAGE_STATUS_OPTIONS = [
  { value: "active", label: "进行中" },
  { value: "done", label: "已完成" },
  { value: "archived", label: "归档" },
];

function Dashboard() {
  const [now, setNow] = useState(() => new Date());
  const [dailyTasks, setDailyTasks] = useState([]);
  const [unscheduledTasks, setUnscheduledTasks] = useState([]);
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

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  const loadPlanning = () => {
    Promise.all([
      listPlanningDailyTasks({ date: today }).catch(() => []),
      listPlanningDailyTasks({ unscheduled: true }).catch(() => []),
      listWeeklyTasks({ week_start: weekStart }).catch(() => []),
      listStageGoals().catch(() => []),
    ]).then(([daily, unscheduled, weekly, stages]) => {
      setDailyTasks(Array.isArray(daily) ? daily : []);
      setUnscheduledTasks(Array.isArray(unscheduled) ? unscheduled : []);
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

  const doneCount = dailyTasks.filter((task) => task.done).length;
  const totalCount = dailyTasks.length;
  const pendingCount = totalCount - doneCount;
  const completionPercent = totalCount === 0 ? 0 : Math.round((doneCount / totalCount) * 100);
  const focusMinutes = logStats?.total_minutes ?? pomodoroStats?.focus_minutes ?? 0;
  const focusCount = pomodoroStats?.focus_count ?? 0;
  const weeklyDoneCount = weeklyTasks.filter((task) => task.status === "done" || task.progress_percent >= 100).length;
  const weeklyProgressPercent = weeklyTasks.length
    ? Math.round(weeklyTasks.reduce((sum, task) => sum + (task.progress_percent || 0), 0) / weeklyTasks.length)
    : 0;
  const activeStageCount = stageGoals.filter((goal) => goal.status !== "done" && goal.status !== "archived").length;
  const stageProgressPercent = stageGoals.length
    ? Math.round(stageGoals.reduce((sum, goal) => sum + (goal.progress_percent || 0), 0) / stageGoals.length)
    : 0;

  const addDailyFast = async () => {
    const title = dailyTitle.trim();
    if (!title) return;
    try {
      await createPlanningDailyTask({ title, plan_date: today, priority: "normal" });
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
      await createWeeklyTask({ title, week_start: weekStart, priority: "normal", status: "todo" });
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
      await createStageGoal({ title, start_date: today, status: "active" });
      setStageTitle("");
      loadPlanning();
      message.success("已添加阶段目标");
    } catch {
      message.error("添加阶段目标失败");
    }
  };

  const toggleDaily = async (task) => {
    setDailyTasks((prev) => prev.map((item) => (item.id === task.id ? { ...item, done: !item.done } : item)));
    try {
      await togglePlanningDailyTask(task.id);
      loadPlanning();
    } catch {
      setDailyTasks((prev) => prev.map((item) => (item.id === task.id ? task : item)));
      message.error("更新任务状态失败");
    }
  };

  const openDetail = (type, item = null) => {
    setDetailModal({ open: true, type, item });
    form.resetFields();
    form.setFieldsValue(toFormValues(type, item, { today, weekStart }));
  };

  const closeDetail = () => setDetailModal({ open: false, type: "daily", item: null });

  const saveDetail = async () => {
    const values = await form.validateFields();
    const payload = fromFormValues(detailModal.type, values);
    try {
      if (detailModal.type === "daily") {
        if (detailModal.item) {
          await updatePlanningDailyTask(detailModal.item.id, payload);
        } else {
          await createPlanningDailyTask(payload);
        }
      }
      if (detailModal.type === "weekly") {
        if (detailModal.item) {
          await updateWeeklyTask(detailModal.item.id, payload);
        } else {
          await createWeeklyTask(payload);
        }
      }
      if (detailModal.type === "stage") {
        if (detailModal.item) {
          await updateStageGoal(detailModal.item.id, payload);
        } else {
          await createStageGoal(payload);
        }
      }
      closeDetail();
      loadPlanning();
      message.success("已保存");
    } catch {
      message.error("保存失败");
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

  const tabs = [
    {
      key: "today",
      label: "日计划",
      children: (
        <div className="dashboard-plan-layout dashboard-plan-layout-today">
          <section className="dashboard-plan-panel dashboard-plan-panel-main">
            <PanelHeader
              icon={<CalendarOutlined />}
              eyebrow="Daily Plan"
              title="今天真正要完成的事"
              meta={`${pendingCount} 项待办`}
              action={<Button type="primary" icon={<PlusOutlined />} onClick={() => openDetail("daily")}>新建日任务</Button>}
            />
            <div className="dashboard-plan-content">
              <TaskQuickInput
                placeholder="输入今日任务，回车快速添加"
                value={dailyTitle}
                onChange={setDailyTitle}
                onSubmit={addDailyFast}
              />
              <DailyTaskList tasks={dailyTasks} onToggle={toggleDaily} onEdit={(task) => openDetail("daily", task)} onDelete={(task) => removeItem("daily", task)} />
            </div>
          </section>
          <aside className="dashboard-plan-side">
            <section className="dashboard-plan-panel">
              <PanelHeader icon={<ClockCircleOutlined />} eyebrow="Today" title="今日节奏" />
              <div className="dashboard-today-summary">
                <div className="dashboard-progress-ring" style={{ "--progress": `${completionPercent}%` }}>
                  <strong>{completionPercent}%</strong>
                  <span>完成率</span>
                </div>
                <div className="dashboard-summary-stack">
                  <MetricPill label="专注时长" value={formatMinutes(focusMinutes)} />
                  <MetricPill label="番茄钟" value={`${focusCount}`} />
                  <MetricPill label="主修科目" value={logStats?.main_subject || "-"} />
                </div>
              </div>
              <RecentLogs logs={recentLogs} />
            </section>
            <section className="dashboard-plan-panel dashboard-plan-panel-quiet">
              <PanelHeader
                icon={<UnorderedListOutlined />}
                eyebrow="Inbox"
                title="待安排池"
                meta={`${unscheduledTasks.length} 项`}
                action={<Button icon={<PlusOutlined />} onClick={() => openDetail("daily")}>添加</Button>}
              />
              <div className="dashboard-plan-content">
                <DailyTaskList compact tasks={unscheduledTasks} onToggle={toggleDaily} onEdit={(task) => openDetail("daily", task)} onDelete={(task) => removeItem("daily", task)} empty="没有待安排任务" />
              </div>
            </section>
          </aside>
        </div>
      ),
    },
    {
      key: "week",
      label: "周计划",
      children: (
        <div className="dashboard-plan-layout dashboard-plan-layout-week">
          <section className="dashboard-plan-panel dashboard-plan-panel-main">
            <PanelHeader
              icon={<FieldTimeOutlined />}
              eyebrow={weekStart}
              title="本周推进列表"
              meta={`${weeklyDoneCount} / ${weeklyTasks.length} 已完成`}
              action={<Button type="primary" icon={<PlusOutlined />} onClick={() => openDetail("weekly")}>新建周任务</Button>}
            />
            <div className="dashboard-plan-content">
              <TaskQuickInput
                placeholder="输入周任务，回车快速添加"
                value={weeklyTitle}
                onChange={setWeeklyTitle}
                onSubmit={addWeeklyFast}
              />
              <WeeklyTaskList tasks={weeklyTasks} onEdit={(task) => openDetail("weekly", task)} onDelete={(task) => removeItem("weekly", task)} />
            </div>
          </section>
          <aside className="dashboard-plan-side">
            <section className="dashboard-plan-panel dashboard-week-brief">
              <PanelHeader icon={<CheckCircleOutlined />} eyebrow="Weekly Review" title="周计划概况" />
              <Progress percent={weeklyProgressPercent} strokeColor={{ from: "#2d7ff9", to: "#2fbf71" }} />
              <div className="dashboard-week-stats">
                <MetricPill label="周任务" value={`${weeklyTasks.length}`} />
                <MetricPill label="已完成" value={`${weeklyDoneCount}`} />
                <MetricPill label="待拆解" value={`${weeklyTasks.filter((task) => !task.daily_total).length}`} />
              </div>
            </section>
            <section className="dashboard-plan-panel dashboard-plan-panel-quiet">
              <PanelHeader icon={<CalendarOutlined />} eyebrow="Breakdown" title="拆到每天" />
              <div className="dashboard-plan-note">周任务负责方向，日任务负责执行。给日任务设置计划执行日后，它会进入当天列表；只设置 DDL 时会留在待安排池。</div>
              <Button block type="primary" icon={<PlusOutlined />} className="dashboard-plan-action" onClick={() => openDetail("daily")}>新增日任务</Button>
            </section>
          </aside>
        </div>
      ),
    },
    {
      key: "stage",
      label: "长期规划",
      children: (
        <section className="dashboard-plan-panel dashboard-stage-panel">
          <PanelHeader
            icon={<FlagOutlined />}
            eyebrow="Long Range"
            title="从考试目标倒推阶段安排"
            meta={`${activeStageCount} 个进行中`}
            action={<Button type="primary" icon={<PlusOutlined />} onClick={() => openDetail("stage")}>新建长期目标</Button>}
          />
          <div className="dashboard-stage-overview">
            <div>
              <span>整体推进</span>
              <strong>{stageProgressPercent}%</strong>
            </div>
            <Progress percent={stageProgressPercent} showInfo={false} strokeColor={{ from: "#2d7ff9", to: "#f59f00" }} />
          </div>
          <TaskQuickInput
            placeholder="输入长期目标，例如：考前两周申论每周 3 篇"
            value={stageTitle}
            onChange={setStageTitle}
            onSubmit={addStageFast}
          />
          <StageGoalList goals={stageGoals} onEdit={(goal) => openDetail("stage", goal)} onDelete={(goal) => removeItem("stage", goal)} />
        </section>
      ),
    },
  ];

  return (
    <div className="page-grid dashboard-page">
      <Card className="glass-card dashboard-hero" bordered={false}>
        <Row gutter={[24, 16]} align="middle">
          <Col xs={24} md={12}>
            <div className="dashboard-hero-time">{formatTime(now)}<span className="dashboard-hero-seconds">{formatSeconds(now)}</span></div>
            <div className="dashboard-hero-date">{formatLongDate(now)} · {weekdayLabel(now)}</div>
          </Col>
          <Col xs={24} md={12}>
            <div className="dashboard-hero-progress">
              <div className="dashboard-hero-progress-head"><span>今日完成</span><strong>{doneCount} / {totalCount || 0}</strong></div>
              <Progress percent={completionPercent} showInfo={false} strokeColor={{ from: "#2d7ff9", to: "#8a76f3" }} />
              <div className="dashboard-hero-progress-foot"><span>未完成 {pendingCount}</span><span>专注 {formatMinutes(focusMinutes)}</span><span>{focusCount} 个番茄钟</span></div>
            </div>
          </Col>
        </Row>
      </Card>

      <Tabs className="dashboard-tabs app-section-tabs" items={tabs} />
      <PlanningModal modal={detailModal} form={form} stageGoals={stageGoals} weeklyTasks={weeklyTasks} onCancel={closeDetail} onSave={saveDetail} />
    </div>
  );
}

function PanelHeader({ icon, eyebrow, title, meta, action }) {
  return (
    <div className="dashboard-panel-header">
      <div className="dashboard-panel-title-group">
        <span className="dashboard-panel-icon">{icon}</span>
        <div>
          <div className="dashboard-panel-eyebrow">{eyebrow}</div>
          <h3>{title}</h3>
        </div>
      </div>
      <div className="dashboard-panel-tools">
        {meta && <span className="dashboard-panel-meta">{meta}</span>}
        {action}
      </div>
    </div>
  );
}

function MetricPill({ label, value }) {
  return (
    <div className="dashboard-metric-pill">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function TaskQuickInput({ value, onChange, onSubmit, placeholder }) {
  return (
    <div className="dashboard-task-input">
      <PlusOutlined className="dashboard-task-input-icon" />
      <Input value={value} onChange={(event) => onChange(event.target.value)} onPressEnter={onSubmit} placeholder={placeholder} allowClear variant="borderless" />
      <span className="dashboard-task-input-hint">Enter</span>
    </div>
  );
}

function DailyTaskList({ tasks, onToggle, onEdit, onDelete, compact = false, empty = "暂无任务" }) {
  if (!tasks.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={empty} />;
  return (
    <List
      className={compact ? "dashboard-task-list compact" : "dashboard-task-list"}
      dataSource={tasks}
      renderItem={(task) => (
        <List.Item className={task.done ? "dashboard-task-item done" : "dashboard-task-item"} actions={[
          <Tooltip title="编辑" key="edit"><Button type="text" size="small" icon={<EditOutlined />} onClick={() => onEdit(task)} /></Tooltip>,
          <Tooltip title="删除" key="delete"><Button type="text" size="small" icon={<DeleteOutlined />} onClick={() => onDelete(task)} /></Tooltip>,
        ]}>
          <Space align="start" size={12} style={{ width: "100%" }}>
            <span className="dashboard-priority-dot" style={{ background: PRIORITY_COLOR[task.priority] || PRIORITY_COLOR.normal }} />
            <Checkbox checked={task.done} onChange={() => onToggle(task)} />
            <div className="dashboard-task-main">
              <div className="dashboard-task-title-row">
                <span className="dashboard-task-title">{task.title}</span>
                <Tag color={PRIORITY_COLOR[task.priority] || PRIORITY_COLOR.normal} bordered={false}>{priorityLabel(task.priority)}</Tag>
              </div>
              <TaskMeta task={task} />
            </div>
          </Space>
        </List.Item>
      )}
    />
  );
}

function WeeklyTaskList({ tasks, onEdit, onDelete }) {
  if (!tasks.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本周还没有任务" />;
  return (
    <List className="dashboard-task-list" dataSource={tasks} renderItem={(task) => (
      <List.Item className="dashboard-week-item" actions={[
        <Tooltip title="编辑" key="edit"><Button type="text" size="small" icon={<EditOutlined />} onClick={() => onEdit(task)} /></Tooltip>,
        <Tooltip title="删除" key="delete"><Button type="text" size="small" icon={<DeleteOutlined />} onClick={() => onDelete(task)} /></Tooltip>,
      ]}>
        <div className="dashboard-task-main">
          <div className="dashboard-week-head">
            <div>
              <Tag color={PRIORITY_COLOR[task.priority] || PRIORITY_COLOR.normal} bordered={false}>{priorityLabel(task.priority)}</Tag>
              <strong>{task.title}</strong>
            </div>
            <Tag color={STATUS_COLOR[task.status] || "default"} bordered={false}>{statusLabel(task.status)}</Tag>
          </div>
          <Progress percent={task.progress_percent || 0} size="small" strokeColor={{ from: "#2d7ff9", to: "#2fbf71" }} />
          <div className="dashboard-week-meta">
            <span>拆解 {task.daily_done || 0}/{task.daily_total || 0}</span>
            <span>DDL {formatDate(task.due_date) || "-"}</span>
          </div>
        </div>
      </List.Item>
    )} />
  );
}

function StageGoalList({ goals, onEdit, onDelete }) {
  if (!goals.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有阶段目标" />;
  return (
    <List className="dashboard-stage-list" dataSource={goals} grid={{ gutter: 12, xs: 1, md: 2, xl: 3 }} renderItem={(goal) => (
      <List.Item>
        <Card className="dashboard-stage-card" bordered={false} actions={[
          <EditOutlined key="edit" onClick={() => onEdit(goal)} />,
          <DeleteOutlined key="delete" onClick={() => onDelete(goal)} />,
        ]}>
          <div className="dashboard-stage-card-head">
            <span className="dashboard-stage-title">{goal.title}</span>
            <Tag color={STATUS_COLOR[goal.status] || "processing"} bordered={false}>{stageStatusLabel(goal.status)}</Tag>
          </div>
          <div className="dashboard-stage-dates">
            <CalendarOutlined />
            <span>{formatDate(goal.start_date) || "未设置"} 至 {formatDate(goal.end_date) || "未设置"}</span>
          </div>
          {goal.target_metric && <div className="dashboard-plan-note dashboard-stage-target">目标：{goal.target_metric}</div>}
          <Progress percent={goal.progress_percent || 0} strokeColor={{ from: "#2d7ff9", to: "#f59f00" }} />
          <div className="dashboard-stage-metrics">
            <MetricPill label="周任务" value={`${goal.weekly_done || 0}/${goal.weekly_total || 0}`} />
            <MetricPill label="日任务" value={`${goal.daily_done || 0}/${goal.daily_total || 0}`} />
          </div>
        </Card>
      </List.Item>
    )} />
  );
}

function RecentLogs({ logs }) {
  return (
    <>
      <div className="dashboard-recent-head"><span>最近学习记录</span></div>
      {logs.length === 0 ? <div className="dashboard-recent-empty">今天还没有学习记录</div> : (
        <List className="dashboard-recent-list" dataSource={logs} renderItem={(log) => (
          <List.Item><Space align="center" size={10}><Tag color="processing">{log.subject || log.study_type || "学习"}</Tag><span>{formatLogRange(log)}</span></Space><span className="dashboard-recent-note">{log.note || log.question_type || ""}</span></List.Item>
        )} />
      )}
    </>
  );
}

function PlanningModal({ modal, form, stageGoals, weeklyTasks, onCancel, onSave }) {
  const typeLabel = modal.type === "daily" ? "日任务" : modal.type === "weekly" ? "周任务" : "阶段目标";
  return (
    <Modal open={modal.open} title={`${modal.item ? "编辑" : "新建"}${typeLabel}`} onCancel={onCancel} onOk={onSave} destroyOnHidden>
      <Form form={form} layout="vertical">
        <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}><Input /></Form.Item>
        {modal.type === "daily" && <>
          <Form.Item name="date_range" label="快速日期范围"><DatePicker.RangePicker style={{ width: "100%" }} /></Form.Item>
          <Row gutter={12}><Col span={12}><Form.Item name="plan_date" label="计划执行日"><DatePicker style={{ width: "100%" }} /></Form.Item></Col><Col span={12}><Form.Item name="due_date" label="DDL"><DatePicker style={{ width: "100%" }} /></Form.Item></Col></Row>
          <Row gutter={12}><Col span={12}><Form.Item name="stage_goal_id" label="阶段目标"><Select allowClear options={stageGoals.map((goal) => ({ value: goal.id, label: goal.title }))} /></Form.Item></Col><Col span={12}><Form.Item name="weekly_task_id" label="周任务"><Select allowClear options={weeklyTasks.map((task) => ({ value: task.id, label: task.title }))} /></Form.Item></Col></Row>
          <Form.Item name="priority" label="优先级"><Select options={PRIORITY_OPTIONS} /></Form.Item>
        </>}
        {modal.type === "weekly" && <>
          <Row gutter={12}><Col span={12}><Form.Item name="week_start" label="所属周"><DatePicker style={{ width: "100%" }} /></Form.Item></Col><Col span={12}><Form.Item name="due_date" label="DDL"><DatePicker style={{ width: "100%" }} /></Form.Item></Col></Row>
          <Row gutter={12}><Col span={12}><Form.Item name="stage_goal_id" label="阶段目标"><Select allowClear options={stageGoals.map((goal) => ({ value: goal.id, label: goal.title }))} /></Form.Item></Col><Col span={12}><Form.Item name="status" label="状态"><Select options={STATUS_OPTIONS} /></Form.Item></Col></Row>
          <Form.Item name="priority" label="优先级"><Select options={PRIORITY_OPTIONS} /></Form.Item>
        </>}
        {modal.type === "stage" && <>
          <Form.Item name="date_range" label="阶段时间"><DatePicker.RangePicker style={{ width: "100%" }} /></Form.Item>
          <Row gutter={12}><Col span={12}><Form.Item name="start_date" label="开始日期"><DatePicker style={{ width: "100%" }} /></Form.Item></Col><Col span={12}><Form.Item name="end_date" label="结束/考前 DDL"><DatePicker style={{ width: "100%" }} /></Form.Item></Col></Row>
          <Form.Item name="target_metric" label="目标指标"><Input placeholder="例如：行测稳定 70+，申论每周 3 篇" /></Form.Item>
          <Form.Item name="status" label="状态"><Select options={STAGE_STATUS_OPTIONS} /></Form.Item>
        </>}
        <Form.Item name="note" label="备注"><Input.TextArea rows={3} /></Form.Item>
      </Form>
    </Modal>
  );
}

function TaskMeta({ task }) {
  const meta = [];
  if (task.plan_date) meta.push(`计划 ${formatDate(task.plan_date)}`);
  if (task.due_date) meta.push(`DDL ${formatDate(task.due_date)}`);
  if (task.note) meta.push(task.note);
  if (!meta.length) return null;
  return <div className="dashboard-task-meta-line">{meta.join(" · ")}</div>;
}

function toFormValues(type, item, fallback) {
  if (!item) {
    if (type === "daily") return { priority: "normal", plan_date: dateObject(fallback.today) };
    if (type === "weekly") return { priority: "normal", status: "todo", week_start: dateObject(fallback.weekStart) };
    return { status: "active", start_date: dateObject(fallback.today) };
  }
  return Object.fromEntries(Object.entries({
    ...item,
    plan_date: dateObject(item.plan_date),
    due_date: dateObject(item.due_date),
    week_start: dateObject(item.week_start),
    start_date: dateObject(item.start_date),
    end_date: dateObject(item.end_date),
  }).filter(([, value]) => value !== undefined));
}

function fromFormValues(type, values) {
  const payload = { ...values };
  delete payload.date_range;
  for (const key of ["plan_date", "due_date", "week_start", "start_date", "end_date"]) payload[key] = formatMaybeDate(payload[key]);
  if (values.date_range?.length === 2) {
    if (type === "daily") {
      payload.plan_date = formatMaybeDate(values.date_range[0]);
      payload.due_date = formatMaybeDate(values.date_range[1]);
    }
    if (type === "stage") {
      payload.start_date = formatMaybeDate(values.date_range[0]);
      payload.end_date = formatMaybeDate(values.date_range[1]);
    }
  }
  return payload;
}

function priorityLabel(priority) {
  return PRIORITY_OPTIONS.find((option) => option.value === priority)?.label || "中";
}

function statusLabel(status) {
  return STATUS_OPTIONS.find((option) => option.value === status)?.label || status || "-";
}

function stageStatusLabel(status) {
  return STAGE_STATUS_OPTIONS.find((option) => option.value === status)?.label || status || "进行中";
}

function formatMinutes(minutes) {
  if (!minutes) return "0m";
  const hours = Math.floor(minutes / 60);
  const rest = minutes % 60;
  if (hours === 0) return `${rest}m`;
  if (rest === 0) return `${hours}h`;
  return `${hours}h ${rest}m`;
}

function formatTime(date) {
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function formatSeconds(date) {
  return `:${String(date.getSeconds()).padStart(2, "0")}`;
}

function formatLongDate(date) {
  return `${date.getFullYear()} 年 ${date.getMonth() + 1} 月 ${date.getDate()} 日`;
}

function weekdayLabel(date) {
  return ["周日", "周一", "周二", "周三", "周四", "周五", "周六"][date.getDay()];
}

function formatDateKey(date) {
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function startOfWeek(date) {
  const result = new Date(date);
  const day = result.getDay() || 7;
  result.setDate(result.getDate() - day + 1);
  return result;
}

function formatDate(value) {
  if (!value) return "";
  return formatDateKey(new Date(value));
}

function dateObject(value) {
  if (!value) return undefined;
  return dayjs(value);
}

function formatMaybeDate(value) {
  if (!value) return "";
  return value.format ? value.format("YYYY-MM-DD") : formatDateKey(new Date(value));
}

function formatLogRange(log) {
  if (!log) return "";
  const start = log.start_time ? new Date(log.start_time) : null;
  const end = log.end_time ? new Date(log.end_time) : null;
  if (start && end) return `${formatTime(start)} - ${formatTime(end)}`;
  if (log.duration_min) return `${log.duration_min} 分钟`;
  return "";
}

export default Dashboard;
