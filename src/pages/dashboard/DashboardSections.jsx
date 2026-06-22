import { Button, Col, Progress, Row } from "antd";
import {
  BulbOutlined,
  CalendarOutlined,
  CheckCircleOutlined,
  FieldTimeOutlined,
  FlagOutlined,
  HistoryOutlined,
  LineChartOutlined,
  PlusOutlined,
  UnorderedListOutlined,
} from "@ant-design/icons";
import { AppCard, SectionHeader, StatCard } from "../../components/ui";
import {
  DailyTaskList,
  DeadlineReminderList,
  InboxTaskList,
  RecentLogs,
  StageGoalList,
  TaskQuickInput,
  WeeklyTaskList,
} from "./DashboardComponents";
import {
  formatLongDate,
  formatMinutes,
  formatTime,
  weekdayLabel,
} from "./dashboardUtils";

// 主题色双色进度条（随主题切换）。
const PROGRESS_COLOR = { from: "var(--color-brand)", to: "var(--color-accent)" };

function DashboardGreeting({ now }) {
  return (
    <div className="dashboard-greeting">
      <strong>{getGreeting(now)}，继续加油吧！</strong>
      <span>{formatLongDate(now)} · {weekdayLabel(now)} · {formatTime(now)}</span>
    </div>
  );
}

function ProgressHint({ percent, text }) {
  return (
    <div className="dashboard-stat-progress">
      <Progress percent={percent} showInfo={false} size="small" strokeColor={PROGRESS_COLOR} />
      {text && <span>{text}</span>}
    </div>
  );
}

export function TodayTab({
  now,
  dailyTasks,
  dailyTitle,
  doneCount,
  totalCount,
  pendingCount,
  completionPercent,
  focusMinutes,
  focusCount,
  mainSubject,
  recentLogs,
  todayDeadlineItems,
  onDailyTitleChange,
  onAddDaily,
  onToggleDaily,
  onEditDaily,
  onDeleteDaily,
  onEditWeekly,
  onArrangeWeekly,
  onCreateDaily,
}) {
  return (
    <div className="dashboard-today-stack">
      <DashboardGreeting now={now} />
      <Row gutter={[16, 16]}>
        <Col xs={12} md={6}>
          <StatCard
            label="今日完成"
            value={`${doneCount} / ${totalCount || 0}`}
            hint={<ProgressHint percent={completionPercent} text={`${completionPercent}% · 未完成 ${pendingCount}`} />}
          />
        </Col>
        <Col xs={12} md={6}><StatCard label="专注时长" value={formatMinutes(focusMinutes)} hint="今日累计" /></Col>
        <Col xs={12} md={6}><StatCard label="番茄" value={`${focusCount}`} hint="完成段数" /></Col>
        <Col xs={12} md={6}><StatCard label="主修" value={mainSubject || "-"} hint="按时长统计" /></Col>
      </Row>
      <div className="dashboard-plan-layout dashboard-plan-layout-today">
        <AppCard className="dashboard-plan-panel dashboard-plan-panel-main dashboard-today-tasks-panel">
          <SectionHeader
            icon={<CalendarOutlined />}
            eyebrow="执行层"
            title="今天真正要完成的事"
            meta={`${pendingCount} 项待办`}
            action={<Button type="primary" icon={<PlusOutlined />} onClick={onCreateDaily}>新建日任务</Button>}
          />
          <div className="dashboard-plan-content">
            <TaskQuickInput
              placeholder="输入今日任务，回车快速添加"
              value={dailyTitle}
              onChange={onDailyTitleChange}
              onSubmit={onAddDaily}
            />
            <DailyTaskList
              tasks={dailyTasks}
              onToggle={onToggleDaily}
              onEdit={onEditDaily}
              onDelete={onDeleteDaily}
              empty="先添加 1 个今天必须完成的小任务"
            />
          </div>
        </AppCard>
        <aside className="dashboard-plan-side dashboard-today-side">
          <AppCard className="dashboard-plan-panel dashboard-compact-panel">
            <SectionHeader
              icon={<UnorderedListOutlined />}
              eyebrow="提醒"
              title="今日到期"
              meta={`${todayDeadlineItems.length} 项`}
            />
            <div className="dashboard-plan-content">
              <DeadlineReminderList
                items={todayDeadlineItems}
                onEditDaily={onEditDaily}
                onEditWeekly={onEditWeekly}
                onArrangeWeekly={onArrangeWeekly}
              />
            </div>
          </AppCard>
          <AppCard className="dashboard-plan-panel dashboard-recent-panel dashboard-compact-panel">
            <SectionHeader icon={<HistoryOutlined />} eyebrow="记录" title="学习记录" meta={`${recentLogs.length} 条`} />
            <RecentLogs logs={recentLogs} />
          </AppCard>
        </aside>
      </div>
      <TodayInsightRow
        pendingCount={pendingCount}
        deadlineCount={todayDeadlineItems.length}
        recentLogCount={recentLogs.length}
        completionPercent={completionPercent}
      />
    </div>
  );
}

function TodayInsightRow({ pendingCount, deadlineCount, recentLogCount, completionPercent }) {
  const nextStep = getTodayNextStep({ pendingCount, deadlineCount, recentLogCount, completionPercent });
  return (
    <div className="dashboard-insight-row">
      <InsightCard icon={<BulbOutlined />} title="下一步" text={nextStep} />
      <InsightCard
        icon={<CalendarOutlined />}
        title="到期检查"
        text={deadlineCount ? `${deadlineCount} 项今天到期，先处理时间敏感项` : "今天暂无到期，优先推进当前待办"}
      />
      <InsightCard
        icon={<LineChartOutlined />}
        title="学习记录"
        text={recentLogCount ? `已有 ${recentLogCount} 条记录，收尾补充错题与备注` : "完成一段学习后补记，方便复盘"}
      />
    </div>
  );
}

function InsightCard({ icon, title, text }) {
  return (
    <div className="dashboard-insight-card">
      <span className="dashboard-insight-icon">{icon}</span>
      <div>
        <strong>{title}</strong>
        <span>{text}</span>
      </div>
    </div>
  );
}

function getTodayNextStep({ pendingCount, deadlineCount, recentLogCount, completionPercent }) {
  if (pendingCount > 0) return `还有 ${pendingCount} 项待办，先完成最小的一项`;
  if (deadlineCount > 0) return "待办已清空，检查到期项是否需要补充安排";
  if (recentLogCount === 0) return "任务为空时，补一条学习记录形成复盘闭环";
  if (completionPercent >= 100) return "今天执行层已收尾，可以安排明天第一项";
  return "从今日任务里选 1 项，拆到下一段专注时间";
}

export function WeekTab({
  weekStart,
  weekEnd,
  weeklyTitle,
  weeklyTasks,
  stageLinkedWeeklyTasks,
  independentWeeklyTasks,
  weeklyInboxTasks,
  inboxTasks,
  stageGoals,
  weeklyProgressPercent,
  weeklyDoneCount,
  onWeeklyTitleChange,
  onAddWeekly,
  onCreateWeekly,
  onEditWeekly,
  onDeleteWeekly,
  onArrangeWeekly,
  onToggleDaily,
  onEditDaily,
  onDeleteDaily,
  onCreateInboxDaily,
}) {
  const weeklyToBreakDownCount = weeklyTasks.filter((task) => !task.daily_total).length;
  const pendingWeeklyCount = Math.max(weeklyTasks.length - weeklyDoneCount, 0);

  return (
    <div className="dashboard-week-stack">
      <Row gutter={[16, 16]}>
        <Col xs={12} md={6}>
          <StatCard label="本周完成" value={`${weeklyProgressPercent}%`} hint={<ProgressHint percent={weeklyProgressPercent} />} />
        </Col>
        <Col xs={12} md={6}><StatCard label="本周任务" value={`${weeklyTasks.length}`} hint={`${weekStart} ~ ${weekEnd}`} /></Col>
        <Col xs={12} md={6}><StatCard label="已完成" value={`${weeklyDoneCount}`} hint="本周达成" /></Col>
        <Col xs={12} md={6}><StatCard label="待拆解" value={`${weeklyToBreakDownCount}`} hint="尚未排期" /></Col>
      </Row>
      <div className="dashboard-plan-layout dashboard-plan-layout-week">
        <AppCard className="dashboard-plan-panel dashboard-plan-panel-main">
          <SectionHeader
            icon={<FieldTimeOutlined />}
            eyebrow="组织层"
            title="绑定长期目标的推进项"
            meta={`${stageLinkedWeeklyTasks.length} 项`}
            action={<Button type="primary" icon={<PlusOutlined />} onClick={onCreateWeekly}>新建周任务</Button>}
          />
          <div className="dashboard-plan-content">
            <TaskQuickInput
              placeholder="输入周任务，回车快速添加"
              value={weeklyTitle}
              onChange={onWeeklyTitleChange}
              onSubmit={onAddWeekly}
            />
            <WeeklyTaskList
              tasks={stageLinkedWeeklyTasks}
              stageGoals={stageGoals}
              onEdit={onEditWeekly}
              onDelete={onDeleteWeekly}
              onArrangeToday={onArrangeWeekly}
              empty="先把长期目标拆成 1-3 个本周推进项"
            />
          </div>
        </AppCard>
        <aside className="dashboard-plan-side">
          <AppCard className="dashboard-plan-panel">
            <SectionHeader icon={<CalendarOutlined />} eyebrow="独立任务" title="本周独立任务" meta={`${independentWeeklyTasks.length} 项`} />
            <WeeklyTaskList
              compact
              tasks={independentWeeklyTasks}
              stageGoals={stageGoals}
              onEdit={onEditWeekly}
              onDelete={onDeleteWeekly}
              onArrangeToday={onArrangeWeekly}
              empty="没有独立任务；临时事项可直接新建周任务"
            />
          </AppCard>
          <AppCard className="dashboard-plan-panel">
            <SectionHeader
              icon={<UnorderedListOutlined />}
              eyebrow="待安排"
              title="待安排池"
              meta={`${inboxTasks.length + weeklyInboxTasks.length} 项`}
              action={<Button icon={<PlusOutlined />} onClick={onCreateInboxDaily}>添加</Button>}
            />
            <InboxTaskList
              dailyTasks={inboxTasks}
              weeklyTasks={weeklyInboxTasks}
              onToggleDaily={onToggleDaily}
              onEditDaily={onEditDaily}
              onDeleteDaily={onDeleteDaily}
              onEditWeekly={onEditWeekly}
              onDeleteWeekly={onDeleteWeekly}
              onArrangeWeekly={onArrangeWeekly}
            />
          </AppCard>
        </aside>
      </div>
      <div className="dashboard-insight-row">
        <InsightCard
          icon={<FlagOutlined />}
          title="绑定长期目标"
          text={stageLinkedWeeklyTasks.length ? `${stageLinkedWeeklyTasks.length} 项正在承接长期目标` : "本周还没有承接长期目标的推进项"}
        />
        <InsightCard
          icon={<CalendarOutlined />}
          title="独立任务"
          text={independentWeeklyTasks.length ? `${independentWeeklyTasks.length} 项不绑定长期目标，适合作为临时推进` : "暂无独立任务，周计划更聚焦"}
        />
        <InsightCard
          icon={<CheckCircleOutlined />}
          title="下一步"
          text={getWeekNextStep({ weeklyTaskCount: weeklyTasks.length, weeklyToBreakDownCount, pendingWeeklyCount })}
        />
      </div>
    </div>
  );
}

function getWeekNextStep({ weeklyTaskCount, weeklyToBreakDownCount, pendingWeeklyCount }) {
  if (!weeklyTaskCount) return "先新建 1 个周任务，决定它是否绑定长期目标";
  if (weeklyToBreakDownCount > 0) return `${weeklyToBreakDownCount} 项待拆解，优先安排到具体日期`;
  if (pendingWeeklyCount > 0) return `还有 ${pendingWeeklyCount} 项待推进，今天挑 1 项落地`;
  return "本周任务已完成，补充复盘或准备下周推进项";
}

export function StageTab({
  stageTitle,
  stageGoals,
  activeStageCount,
  completedStageCount,
  stageProgressPercent,
  onStageTitleChange,
  onAddStage,
  onCreateStage,
  onEditStage,
  onDeleteStage,
}) {
  return (
    <div className="dashboard-stage-stack">
      <Row gutter={[16, 16]}>
        <Col xs={12} md={6}>
          <StatCard label="整体推进" value={`${stageProgressPercent}%`} hint={<ProgressHint percent={stageProgressPercent} />} />
        </Col>
        <Col xs={12} md={6}><StatCard label="目标总数" value={`${stageGoals.length}`} hint="长期目标" /></Col>
        <Col xs={12} md={6}><StatCard label="进行中" value={`${activeStageCount}`} hint="待推进" /></Col>
        <Col xs={12} md={6}><StatCard label="已完成" value={`${completedStageCount}`} hint="已达成" /></Col>
      </Row>
      <AppCard className="dashboard-plan-panel dashboard-stage-panel">
        <SectionHeader
          icon={<FlagOutlined />}
          eyebrow="目标层"
          title="长期目标列表"
          meta={`${stageGoals.length} 个目标`}
          action={<Button type="primary" icon={<PlusOutlined />} onClick={onCreateStage}>新建长期目标</Button>}
        />
        <TaskQuickInput
          placeholder="输入长期目标，例如：考前两周申论每周 3 篇"
          value={stageTitle}
          onChange={onStageTitleChange}
          onSubmit={onAddStage}
        />
        <StageGoalList goals={stageGoals} onEdit={onEditStage} onDelete={onDeleteStage} />
      </AppCard>
      <div className="dashboard-insight-row">
        <InsightCard
          icon={<LineChartOutlined />}
          title="阶段推进"
          text={stageGoals.length ? `整体推进 ${stageProgressPercent}%，优先更新进行中目标` : "还没有长期目标，先建立一个可拆解目标"}
        />
        <InsightCard
          icon={<FieldTimeOutlined />}
          title="下一阶段"
          text={activeStageCount ? `${activeStageCount} 个目标进行中，选择一个明确下一阶段成果` : "暂无进行中目标，先新建或恢复一个目标"}
        />
        <InsightCard
          icon={<CalendarOutlined />}
          title="拆到周计划"
          text={getStageNextStep({ stageCount: stageGoals.length, activeStageCount, completedStageCount })}
        />
      </div>
    </div>
  );
}

function getStageNextStep({ stageCount, activeStageCount, completedStageCount }) {
  if (!stageCount) return "先新建长期目标，再到周计划里绑定推进项";
  if (activeStageCount > 0) return "从进行中目标挑 1 个，拆成下周 1-3 个推进项";
  if (completedStageCount > 0) return "已完成目标可复盘，再开启下一个阶段";
  return "给目标设置状态和阶段成果，再拆到周计划";
}

function getGreeting(now) {
  const hour = now.getHours();
  if (hour < 6) return "夜深了";
  if (hour < 12) return "早安";
  if (hour < 18) return "下午好";
  return "晚上好";
}
