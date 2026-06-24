import { Button } from "antd";
import {
  CalendarOutlined,
  FieldTimeOutlined,
  FlagOutlined,
  HistoryOutlined,
  PlusOutlined,
  UnorderedListOutlined,
} from "@ant-design/icons";
import { AppCard, SectionHeader } from "../../components/ui";
import {
  DailyTaskList,
  DeadlineReminderList,
  InboxTaskList,
  RecentLogs,
  StageGoalList,
  TaskQuickInput,
  WeeklyTaskList,
} from "./DashboardComponents";
import { formatLongDate, weekdayLabel } from "./dashboardUtils";

// 页面顶部主视觉：日期 kicker + 大号问候，取代旧的 PageHeader 页面说明 + 独立问候条。
// 不再显示实时时分（侧边栏时钟卡已有），避免重复、收紧层级。
export function DashboardHero({ now }) {
  return (
    <header className="dashboard-hero">
      <p className="dashboard-hero-eyebrow">{formatLongDate(now)} · {weekdayLabel(now)}</p>
      <h1 className="dashboard-hero-title">{getGreeting(now)}，继续加油吧</h1>
    </header>
  );
}

export function TodayTab({
  dailyTasks,
  dailyTitle,
  pendingCount,
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
    </div>
  );
}

export function WeekTab({
  weeklyTitle,
  stageLinkedWeeklyTasks,
  independentWeeklyTasks,
  weeklyInboxTasks,
  inboxTasks,
  stageGoals,
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
  return (
    <div className="dashboard-week-stack">
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
    </div>
  );
}

export function StageTab({
  stageTitle,
  stageGoals,
  onStageTitleChange,
  onAddStage,
  onCreateStage,
  onEditStage,
  onDeleteStage,
}) {
  return (
    <div className="dashboard-stage-stack">
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
    </div>
  );
}

function getGreeting(now) {
  const hour = now.getHours();
  if (hour < 6) return "夜深了";
  if (hour < 12) return "早安";
  if (hour < 18) return "下午好";
  return "晚上好";
}
