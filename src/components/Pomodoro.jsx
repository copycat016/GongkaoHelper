import { useEffect, useMemo, useState } from "react";
import {
  Button,
  Input,
  InputNumber,
  Cascader,
  message,
  Select,
  Tag,
} from "antd";
import {
  CheckCircleOutlined,
  CoffeeOutlined,
  FireOutlined,
  LeftOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  RightOutlined,
  RightCircleOutlined,
  SettingOutlined,
} from "@ant-design/icons";
import { getTodayPomodoroStats, savePomodoroSession } from "../api/pomodoro";
import { completePlan, getPlans } from "../api/plans";

const PRESETS = [
  { key: "classic", label: "25 / 5", focusMinutes: 25, breakMinutes: 5, desc: "经典番茄钟" },
  { key: "deep", label: "50 / 10", focusMinutes: 50, breakMinutes: 10, desc: "深度学习" },
  { key: "long", label: "90 / 15", focusMinutes: 90, breakMinutes: 15, desc: "长时间沉浸" },
  { key: "light", label: "15 / 3", focusMinutes: 15, breakMinutes: 3, desc: "轻量启动" },
  { key: "custom", label: "自定义", focusMinutes: 25, breakMinutes: 5, desc: "自己设置" },
];

const TASK_TOPIC_OPTIONS = [
  {
    value: "行测刷题",
    label: "行测刷题",
    children: ["常识判断", "言语理解与表达", "数量关系", "判断推理", "资料分析", "图形推理", "定义判断", "类比推理", "逻辑判断"].map((value) => ({ value, label: value })),
  },
  {
    value: "申论练习",
    label: "申论练习",
    children: ["归纳概括题", "综合分析题", "提出对策题", "应用文写作题", "文章论述题", "公文写作题", "材料阅读", "答案复盘"].map((value) => ({ value, label: value })),
  },
  {
    value: "错题复习",
    label: "错题复习",
    children: ["常识错题", "言语错题", "数量错题", "判断错题", "资料错题", "申论错题"].map((value) => ({ value, label: value })),
  },
  {
    value: "学习计划",
    label: "学习计划",
    children: ["每日任务", "每周任务", "阶段任务", "随手任务", "复盘整理"].map((value) => ({ value, label: value })),
  },
  {
    value: "PDF 阅读",
    label: "PDF 阅读",
    children: ["教材阅读", "资料提取", "真题整理", "笔记摘录"].map((value) => ({ value, label: value })),
  },
];

const STORAGE_KEY = "aozora-pomodoro-state";

const defaultPomodoroState = {
  mode: "focus",
  running: false,
  presetKey: "classic",
  durations: { focus: 25, break: 5 },
  customFocus: 25,
  customBreak: 5,
  secondsLeft: 25 * 60,
  taskType: "行测刷题",
  taskName: "资料分析专项",
  taskTopic: "资料分析",
  boundPlanId: undefined,
  timerEndAt: null,
};

function loadPomodoroState() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY));
    if (!saved) return defaultPomodoroState;
    const timerEndAt = saved.timerEndAt || null;
    const secondsLeft = saved.running && timerEndAt
      ? Math.max(0, Math.ceil((timerEndAt - Date.now()) / 1000))
      : Number(saved.secondsLeft || defaultPomodoroState.secondsLeft);
    return {
      ...defaultPomodoroState,
      ...saved,
      durations: { ...defaultPomodoroState.durations, ...(saved.durations || {}) },
      running: Boolean(saved.running && secondsLeft > 0),
      secondsLeft,
      timerEndAt: secondsLeft > 0 ? timerEndAt : null,
    };
  } catch {
    return defaultPomodoroState;
  }
}

function formatTime(totalSeconds) {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

function formatMinutes(minutes) {
  if (!minutes) return "0m";
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

function Pomodoro() {
  const initialState = useMemo(() => loadPomodoroState(), []);
  const [mode, setMode] = useState(initialState.mode);
  const [running, setRunning] = useState(initialState.running);
  const [presetKey, setPresetKey] = useState(initialState.presetKey);
  const [durations, setDurations] = useState(initialState.durations);
  const [customFocus, setCustomFocus] = useState(initialState.customFocus);
  const [customBreak, setCustomBreak] = useState(initialState.customBreak);
  const [secondsLeft, setSecondsLeft] = useState(initialState.secondsLeft);
  const [timerEndAt, setTimerEndAt] = useState(initialState.timerEndAt);
  const [taskType, setTaskType] = useState(initialState.taskType);
  const [taskName, setTaskName] = useState(initialState.taskName);
  const [taskTopic, setTaskTopic] = useState(initialState.taskTopic);
  const [boundPlanId, setBoundPlanId] = useState(initialState.boundPlanId);
  const [dailyPlans, setDailyPlans] = useState([]);
  const [todayStats, setTodayStats] = useState({ focus_count: 0, focus_minutes: 0 });
  const [drawerVisible, setDrawerVisible] = useState(false);

  const isFocus = mode === "focus";
  const totalSeconds = useMemo(() => (isFocus ? durations.focus : durations.break) * 60, [isFocus, durations]);
  const percent = useMemo(() => {
    if (totalSeconds <= 0) return 0;
    const value = ((totalSeconds - secondsLeft) / totalSeconds) * 100;
    return Math.min(100, Math.max(0, Math.round(value)));
  }, [secondsLeft, totalSeconds]);
  const currentPreset = PRESETS.find((item) => item.key === presetKey);

  const loadTodayStats = async () => {
    const stats = await getTodayPomodoroStats();
    setTodayStats(stats || { focus_count: 0, focus_minutes: 0 });
  };

  useEffect(() => {
    loadTodayStats().catch(() => {});
    getPlans()
      .then((items) => setDailyPlans((items || []).filter((item) => item.status !== "已完成" && item.plan_type === "每日计划")))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!running) return;
    if (!timerEndAt) {
      setTimerEndAt(Date.now() + secondsLeft * 1000);
      return;
    }
    const timer = setInterval(() => {
      setSecondsLeft(Math.max(0, Math.ceil((timerEndAt - Date.now()) / 1000)));
    }, 1000);
    setSecondsLeft(Math.max(0, Math.ceil((timerEndAt - Date.now()) / 1000)));
    return () => clearInterval(timer);
  }, [running, secondsLeft, timerEndAt]);

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({
      mode, running, presetKey, durations, customFocus, customBreak,
      secondsLeft, taskType, taskName, taskTopic, boundPlanId, timerEndAt,
    }));
  }, [mode, running, presetKey, durations, customFocus, customBreak, secondsLeft, taskType, taskName, taskTopic, boundPlanId, timerEndAt]);

  useEffect(() => {
    if (secondsLeft !== 0) return;
    setRunning(false);
    setTimerEndAt(null);
    if (mode === "focus") {
      setMode("break");
      setSecondsLeft(durations.break * 60);
    } else {
      setMode("focus");
      setSecondsLeft(durations.focus * 60);
    }
  }, [secondsLeft, mode, durations]);

  const resetToFocus = (nextDurations = durations) => {
    setRunning(false);
    setTimerEndAt(null);
    setMode("focus");
    setSecondsLeft(nextDurations.focus * 60);
  };

  const handlePresetChange = (key) => {
    setPresetKey(key);
    if (key === "custom") {
      const nextDurations = { focus: customFocus, break: customBreak };
      setDurations(nextDurations);
      resetToFocus(nextDurations);
      return;
    }
    const preset = PRESETS.find((item) => item.key === key);
    const nextDurations = { focus: preset.focusMinutes, break: preset.breakMinutes };
    setDurations(nextDurations);
    resetToFocus(nextDurations);
    setDrawerVisible(false);
  };

  const handleApplyCustom = () => {
    const nextDurations = { focus: customFocus, break: customBreak };
    setPresetKey("custom");
    setDurations(nextDurations);
    resetToFocus(nextDurations);
    setDrawerVisible(false);
  };

  const handleStartPause = () => {
    if (running) {
      setRunning(false);
      setTimerEndAt(null);
      return;
    }
    setTimerEndAt(Date.now() + secondsLeft * 1000);
    setRunning(true);
  };

  const handleReset = () => {
    setRunning(false);
    setTimerEndAt(null);
    setSecondsLeft(isFocus ? durations.focus * 60 : durations.break * 60);
  };

  const handleSwitchMode = () => {
    setRunning(false);
    setTimerEndAt(null);
    if (isFocus) {
      setMode("break");
      setSecondsLeft(durations.break * 60);
    } else {
      setMode("focus");
      setSecondsLeft(durations.focus * 60);
    }
  };

  const handleComplete = async () => {
    setRunning(false);
    setTimerEndAt(null);
    const actualMinutes = Math.max(1, Math.ceil((totalSeconds - secondsLeft) / 60));
    await savePomodoroSession({
      task_type: taskType,
      task_name: taskTopic || taskName,
      mode,
      planned_minutes: isFocus ? durations.focus : durations.break,
      actual_minutes: actualMinutes,
      completed_at: new Date().toISOString(),
    });
    if (isFocus && boundPlanId) {
      await completePlan(boundPlanId);
    }
    await loadTodayStats();
    window.dispatchEvent(new Event("pomodoro:updated"));
    message.success("已完成，并写入今日番茄钟记录");
    handleSwitchMode();
  };

  const tipText = isFocus
    ? "专注时间：建议只做一件事，把网页、手机和杂念都先关小声。"
    : "休息时间：站起来、喝水、看远处，给下一轮留一点清醒。";

  return (
    <div className={isFocus ? "pomo-workspace focus" : "pomo-workspace break"}>
      <section className="pomo-stage glass-panel">
        <div className="pomo-stage-main">
          <div className="pomo-stage-summary">
            <Tag icon={isFocus ? <FireOutlined /> : <CoffeeOutlined />} className={isFocus ? "pomo-mode-badge focus" : "pomo-mode-badge break"}>
              {isFocus ? "专注模式" : "休息模式"}
            </Tag>
            <Tag className="pomo-mode-badge neutral">
              今日 {todayStats.focus_count || 0} 轮 · {formatMinutes(todayStats.focus_minutes || 0)}
            </Tag>
            <Tag className="pomo-mode-badge neutral">
              {running ? "计时中" : "待开始"} · {currentPreset?.desc || "自定义"}
            </Tag>
          </div>

          <div className="pomo-timer-wrap">
            <div className="pomo-timer-ring" style={{ "--pomo-progress": `${percent}%` }}>
              <div className="pomo-timer-inner">
                <div className="pomo-timer-label">{isFocus ? "专注" : "休息"}</div>
                <div className="pomo-timer-time">{formatTime(secondsLeft)}</div>
              </div>
            </div>
          </div>
          <div className="pomo-timer-task">{taskType} · {taskTopic || taskName || "未命名任务"}</div>

          <div className="pomo-actions">
            <Button
              type="primary"
              size="large"
              icon={running ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
              onClick={handleStartPause}
              className="pomo-btn-primary"
            >
              {running ? "暂停" : isFocus ? "开始专注" : "开始休息"}
            </Button>
            <Button size="large" icon={<ReloadOutlined />} onClick={handleReset} className="soft-button" title="重置" aria-label="重置" />
            <Button size="large" icon={<RightCircleOutlined />} onClick={handleSwitchMode} className="soft-button">
              {isFocus ? "休息" : "专注"}
            </Button>
            <Button size="large" icon={<CheckCircleOutlined />} onClick={handleComplete} className="soft-button">
              完成
            </Button>
          </div>

          <div className={isFocus ? "pomo-tip focus" : "pomo-tip break"}>
            {tipText}
          </div>
        </div>

        <aside className={drawerVisible ? "pomo-side-panel open" : "pomo-side-panel"}>
          <button
            type="button"
            className="pomo-drawer-handle"
            onClick={() => setDrawerVisible((value) => !value)}
            aria-label={drawerVisible ? "收起设置面板" : "展开设置面板"}
          >
            {drawerVisible ? <RightOutlined /> : <LeftOutlined />}
          </button>

          <div className="pomo-drawer-content">
            <div className="pomo-current-plan">
              <div>
                <span className="pomo-current-label">当前方案</span>
                <strong>{currentPreset?.desc || "自定义"} · {durations.focus}m / {durations.break}m</strong>
              </div>
            </div>

            <div className="pomo-side-block">
              <div className="pomo-side-title">
                <SettingOutlined />
                <span>时间方案</span>
              </div>
              <div className="pomo-preset-grid">
                {PRESETS.map((preset) => (
                  <button
                    key={preset.key}
                    type="button"
                    className={`pomo-preset-item ${presetKey === preset.key ? "active" : ""}`}
                    onClick={() => handlePresetChange(preset.key)}
                  >
                    <span className="pomo-preset-time">{preset.label}</span>
                    <span className="pomo-preset-name">{preset.desc}</span>
                  </button>
                ))}
              </div>
              {presetKey === "custom" && (
                <div className="pomo-custom-row">
                  <div>
                    <div className="pomo-input-label">专注分钟</div>
                    <InputNumber min={1} max={180} value={customFocus} onChange={(value) => setCustomFocus(Number(value || 1))} className="pomo-input" />
                  </div>
                  <div>
                    <div className="pomo-input-label">休息分钟</div>
                    <InputNumber min={1} max={60} value={customBreak} onChange={(value) => setCustomBreak(Number(value || 1))} className="pomo-input" />
                  </div>
                  <Button block icon={<SettingOutlined />} onClick={handleApplyCustom} className="soft-button">
                    应用自定义
                  </Button>
                </div>
              )}
            </div>

            <div className="pomo-side-block pomo-task-block">
              <div className="pomo-side-title">
                <CheckCircleOutlined />
                <span>任务绑定</span>
              </div>
              <div className="pomo-task-form">
                <Select
                  allowClear
                  placeholder="选择每日任务"
                  value={boundPlanId}
                  onChange={(value) => {
                    setBoundPlanId(value);
                    const plan = dailyPlans.find((item) => item.id === value);
                    if (plan) {
                      const nextTaskType = taskTypeFromPlan(plan);
                      setTaskType(nextTaskType);
                      setTaskTopic(plan.question_type || defaultTopicForTask(nextTaskType));
                      setTaskName(plan.title);
                    }
                  }}
                  className="pomo-select"
                  options={dailyPlans.map((item) => ({ value: item.id, label: item.title }))}
                />
                <div className="pomo-task-inline">
                  <Cascader
                    value={[taskType, taskTopic].filter(Boolean)}
                    onChange={(value) => {
                      const [nextTaskType, nextTopic] = value || [];
                      if (nextTaskType) setTaskType(nextTaskType);
                      if (nextTopic) setTaskTopic(nextTopic);
                    }}
                    options={TASK_TOPIC_OPTIONS}
                    placeholder="学习项目"
                    className="pomo-select"
                  />
                  <Input
                    value={taskName}
                    onChange={(event) => setTaskName(event.target.value)}
                    placeholder="补充说明"
                    className="pomo-task-input"
                  />
                </div>
              </div>
            </div>
          </div>
        </aside>
      </section>
    </div>
  );
}

function taskTypeFromPlan(plan) {
  if (plan.subject === "行测") return "行测刷题";
  if (plan.subject === "申论") return "申论练习";
  return "学习计划";
}

function defaultTopicForTask(taskType) {
  const group = TASK_TOPIC_OPTIONS.find((item) => item.value === taskType);
  return group?.children?.[0]?.value || "";
}

export default Pomodoro;
