import {
  Alert,
  Button,
  Card,
  Checkbox,
  Col,
  DatePicker,
  Form,
  Input,
  InputNumber,
  List,
  Modal,
  Progress,
  Row,
  Select,
  Space,
  Tag,
  Tooltip,
} from "antd";
import {
  CalendarOutlined,
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import { EmptyState, MetricPill } from "../../components/ui";
import {
  EXECUTE_MODE_OPTIONS,
  PRIORITY_COLOR,
  PRIORITY_OPTIONS,
  PROGRESS_MODE_OPTIONS,
  STAGE_STATUS_OPTIONS,
  STATUS_COLOR,
  STATUS_OPTIONS,
  executeModeLabel,
  formatDate,
  formatLogRange,
  getDeadlineDate,
  getTargetText,
  getTaskDate,
  isTaskDone,
  priorityLabel,
  progressModeLabel,
  stageStatusLabel,
  stageTitleById,
  statusLabel,
  taskOriginColor,
  taskOriginLabel,
} from "./dashboardUtils";

export function TaskQuickInput({ value, onChange, onSubmit, placeholder }) {
  return (
    <div className="dashboard-task-input">
      <PlusOutlined className="dashboard-task-input-icon" />
      <Input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onPressEnter={onSubmit}
        placeholder={placeholder}
        allowClear
        variant="borderless"
      />
      <span className="dashboard-task-input-hint">Enter</span>
    </div>
  );
}

export function DailyTaskList({ tasks, onToggle, onEdit, onDelete, compact = false, empty = "暂无任务" }) {
  if (!tasks.length) return <EmptyState description={empty} />;
  return (
    <List
      className={compact ? "dashboard-task-list compact" : "dashboard-task-list"}
      dataSource={tasks}
      renderItem={(task) => (
        <List.Item
          className={isTaskDone(task) ? "dashboard-task-item done" : "dashboard-task-item"}
          actions={[
            <Tooltip title="编辑" key="edit">
              <Button type="text" size="small" icon={<EditOutlined />} onClick={() => onEdit(task)} />
            </Tooltip>,
            <Tooltip title="删除" key="delete">
              <Button type="text" size="small" icon={<DeleteOutlined />} onClick={() => onDelete(task)} />
            </Tooltip>,
          ]}
        >
          <Space align="start" size={12} style={{ width: "100%" }}>
            <span className="dashboard-priority-dot" style={{ background: PRIORITY_COLOR[task.priority] || PRIORITY_COLOR.normal }} />
            <Checkbox checked={isTaskDone(task)} onChange={() => onToggle(task)} />
            <div className="dashboard-task-main">
              <div className="dashboard-task-title-row">
                <span className="dashboard-task-title">{task.title}</span>
                <Tag color={PRIORITY_COLOR[task.priority] || PRIORITY_COLOR.normal} bordered={false}>{priorityLabel(task.priority)}</Tag>
                {taskOriginLabel(task) && <Tag color={taskOriginColor(task)} bordered={false}>{taskOriginLabel(task)}</Tag>}
              </div>
              <TaskMeta task={task} />
            </div>
          </Space>
        </List.Item>
      )}
    />
  );
}

export function WeeklyTaskList({ tasks, stageGoals, onEdit, onDelete, onArrangeToday, compact = false, empty = "本周还没有任务" }) {
  if (!tasks.length) return <EmptyState description={empty} />;
  return (
    <List
      className={compact ? "dashboard-task-list compact" : "dashboard-task-list"}
      dataSource={tasks}
      renderItem={(task) => (
        <List.Item
          className="dashboard-week-item"
          actions={[
            <Tooltip title="安排到今天" key="today">
              <Button type="text" size="small" icon={<CalendarOutlined />} onClick={() => onArrangeToday(task)} />
            </Tooltip>,
            <Tooltip title="编辑" key="edit">
              <Button type="text" size="small" icon={<EditOutlined />} onClick={() => onEdit(task)} />
            </Tooltip>,
            <Tooltip title="删除" key="delete">
              <Button type="text" size="small" icon={<DeleteOutlined />} onClick={() => onDelete(task)} />
            </Tooltip>,
          ]}
        >
          <div className="dashboard-task-main">
            <div className="dashboard-week-head">
              <div>
                <Tag color={PRIORITY_COLOR[task.priority] || PRIORITY_COLOR.normal} bordered={false}>{priorityLabel(task.priority)}</Tag>
                <strong>{task.title}</strong>
              </div>
              <Tag color={STATUS_COLOR[task.status] || "default"} bordered={false}>{statusLabel(task.status)}</Tag>
            </div>
            <Progress percent={task.progress_percent || 0} size="small" strokeColor={{ from: "var(--color-brand)", to: "var(--color-accent)" }} />
            <div className="dashboard-week-meta">
              {task.stage_goal_id && <span>阶段：{stageTitleById(stageGoals, task.stage_goal_id)}</span>}
              <span>{executeModeLabel(task.execute_mode)}</span>
              <span>日任务 {task.daily_done || 0}/{task.daily_total || 0}</span>
              <span>DDL {formatDate(getDeadlineDate(task)) || "-"}</span>
            </div>
          </div>
        </List.Item>
      )}
    />
  );
}

export function StageGoalList({ goals, onEdit, onDelete }) {
  if (!goals.length) return <EmptyState description="先创建一个长期目标，再拆成本周推进" />;
  return (
    <div className="dashboard-stage-grid">
      {goals.map((goal) => (
        <Card
          key={goal.id}
          className="dashboard-stage-card"
          bordered={false}
          actions={[
            <EditOutlined key="edit" onClick={() => onEdit(goal)} />,
            <DeleteOutlined key="delete" onClick={() => onDelete(goal)} />,
          ]}
        >
          <div className="dashboard-stage-card-head">
            <span className="dashboard-stage-title">{goal.title}</span>
            <Tag color={STATUS_COLOR[goal.status] || "processing"} bordered={false}>{stageStatusLabel(goal.status)}</Tag>
          </div>
          <div className="dashboard-stage-dates">
            <CalendarOutlined />
            <span>{formatDate(goal.start_date) || "未设置"} 至 {formatDate(goal.end_date) || "未设置"}</span>
          </div>
          {getTargetText(goal) && <div className="dashboard-plan-note dashboard-stage-target">成果：{getTargetText(goal)}</div>}
          <div className="dashboard-stage-card-progress">
            <div>
              <span>目标进度</span>
              <strong>{goal.progress_percent || 0}%</strong>
            </div>
            <Progress percent={goal.progress_percent || 0} showInfo={false} strokeColor={{ from: "var(--color-brand)", to: "var(--color-accent)" }} />
          </div>
          <div className="dashboard-stage-metrics">
            <MetricPill label="周任务" value={`${goal.weekly_done || 0}/${goal.weekly_total || 0}`} />
            <MetricPill label="推进方式" value={progressModeLabel(goal.progress_mode)} />
          </div>
          <CurrentWeeklyList tasks={goal.current_weekly_tasks || []} />
        </Card>
      ))}
    </div>
  );
}

export function DeadlineReminderList({ items, onEditDaily, onEditWeekly, onArrangeWeekly }) {
  if (!items.length) return <EmptyState description="今天没有到期提醒，先推进当前待办" />;
  return (
    <List
      className="dashboard-deadline-list"
      dataSource={items}
      renderItem={({ type, item }) => (
        <List.Item
          actions={[
            type === "weekly" && (
              <Tooltip title="安排到今天" key="today">
                <Button type="text" size="small" icon={<CalendarOutlined />} onClick={() => onArrangeWeekly(item)} />
              </Tooltip>
            ),
            <Tooltip title="编辑" key="edit">
              <Button type="text" size="small" icon={<EditOutlined />} onClick={() => (type === "weekly" ? onEditWeekly(item) : onEditDaily(item))} />
            </Tooltip>,
          ].filter(Boolean)}
        >
          <div className="dashboard-reminder-item">
            <Tag color={type === "weekly" ? "processing" : "warning"} bordered={false}>{type === "weekly" ? "周任务" : "日任务"}</Tag>
            <span>{item.title}</span>
          </div>
        </List.Item>
      )}
    />
  );
}

export function InboxTaskList({
  dailyTasks,
  weeklyTasks,
  onToggleDaily,
  onEditDaily,
  onDeleteDaily,
  onEditWeekly,
  onDeleteWeekly,
  onArrangeWeekly,
}) {
  if (!dailyTasks.length && !weeklyTasks.length) {
    return <EmptyState description="没有待安排事项；需要时添加临时任务或 DDL" />;
  }
  return (
    <div className="dashboard-inbox-stack">
      {!!weeklyTasks.length && (
        <List
          className="dashboard-task-list compact"
          dataSource={weeklyTasks}
          renderItem={(task) => (
            <List.Item
              className="dashboard-week-item"
              actions={[
                <Tooltip title="安排到今天" key="today">
                  <Button type="text" size="small" icon={<CalendarOutlined />} onClick={() => onArrangeWeekly(task)} />
                </Tooltip>,
                <Tooltip title="编辑" key="edit">
                  <Button type="text" size="small" icon={<EditOutlined />} onClick={() => onEditWeekly(task)} />
                </Tooltip>,
                <Tooltip title="删除" key="delete">
                  <Button type="text" size="small" icon={<DeleteOutlined />} onClick={() => onDeleteWeekly(task)} />
                </Tooltip>,
              ]}
            >
              <div className="dashboard-task-main">
                <div className="dashboard-week-head">
                  <div><Tag color="warning" bordered={false}>DDL</Tag><strong>{task.title}</strong></div>
                  <Tag bordered={false}>{executeModeLabel(task.execute_mode)}</Tag>
                </div>
                <div className="dashboard-week-meta"><span>DDL {formatDate(getDeadlineDate(task)) || "-"}</span></div>
              </div>
            </List.Item>
          )}
        />
      )}
      {!!dailyTasks.length && (
        <DailyTaskList
          compact
          tasks={dailyTasks}
          onToggle={onToggleDaily}
          onEdit={onEditDaily}
          onDelete={onDeleteDaily}
          empty="没有待安排事项"
        />
      )}
    </div>
  );
}

export function RecentLogs({ logs }) {
  return (
    <>
      {logs.length === 0 ? (
        <div className="dashboard-recent-empty">还没有学习记录；完成一段学习后补记复盘</div>
      ) : (
        <List
          className="dashboard-recent-list"
          dataSource={logs}
          renderItem={(log) => (
            <List.Item>
              <Space align="center" size={10}>
                <Tag color="processing">{log.subject || log.study_type || "学习"}</Tag>
                <span>{formatLogRange(log)}</span>
              </Space>
              <span className="dashboard-recent-note">{log.note || log.question_type || ""}</span>
            </List.Item>
          )}
        />
      )}
    </>
  );
}

export function PlanningModal({ modal, form, stageGoals, weeklyTasks, onCancel, onSave }) {
  const typeLabel = modal.type === "daily" ? "日任务" : modal.type === "weekly" ? "周任务" : "阶段目标";
  const executeMode = Form.useWatch("execute_mode", form);
  const weekRange = Form.useWatch("week_range", form);
  const validateDateInWeek = (_, value) => {
    if (!value || !weekRange?.[0] || !weekRange?.[1]) return Promise.resolve();
    if (value.isBefore(weekRange[0], "day") || value.isAfter(weekRange[1], "day")) {
      return Promise.reject(new Error("执行日期需在所属周内"));
    }
    return Promise.resolve();
  };
  const validateSplitRange = (_, value) => {
    if (!value?.[0] || !value?.[1]) return Promise.resolve();
    if (value[1].diff(value[0], "day") > 13) {
      return Promise.reject(new Error("最多一次拆解 14 天"));
    }
    if (weekRange?.[0] && weekRange?.[1] && (value[0].isBefore(weekRange[0], "day") || value[1].isAfter(weekRange[1], "day"))) {
      return Promise.reject(new Error("拆解日期需在所属周内"));
    }
    return Promise.resolve();
  };
  return (
    <Modal open={modal.open} title={`${modal.item ? "编辑" : "新建"}${typeLabel}`} onCancel={onCancel} onOk={onSave} destroyOnHidden width={680}>
      <Form form={form} layout="vertical">
        {modal.type === "daily" && modal.item?.origin === "weekly_materialized" && (
          <Alert
            className="dashboard-form-alert"
            type="info"
            showIcon
            message="该日任务由周任务执行计划生成"
            description="可以单独调整它；后续修改周任务执行方式时，只会自动清理未完成的物化任务。"
          />
        )}
        <Form.Item name="title" label="标题" rules={[{ required: true, message: "请输入标题" }]}><Input /></Form.Item>
        {modal.type === "daily" && (
          <>
            <Row gutter={12}>
              <Col xs={24} md={12}><Form.Item name="date" label="执行日期"><DatePicker allowClear style={{ width: "100%" }} /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="deadline" label="DDL"><DatePicker allowClear style={{ width: "100%" }} /></Form.Item></Col>
            </Row>
            <Row gutter={12}>
              <Col xs={24} md={12}><Form.Item name="weekly_task_id" label="来自周任务"><Select allowClear placeholder="可选" options={weeklyTasks.map((task) => ({ value: task.id, label: task.title }))} /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="stage_goal_id" label="关联阶段目标"><Select allowClear placeholder="可选" options={stageGoals.map((goal) => ({ value: goal.id, label: goal.title }))} /></Form.Item></Col>
            </Row>
            <Row gutter={12}>
              <Col xs={24} md={8}><Form.Item name="status" label="状态"><Select options={STATUS_OPTIONS} /></Form.Item></Col>
              <Col xs={24} md={8}><Form.Item name="priority" label="优先级"><Select options={PRIORITY_OPTIONS} /></Form.Item></Col>
              <Col xs={24} md={8}><Form.Item name="estimated_minutes" label="预计分钟"><InputNumber min={0} style={{ width: "100%" }} /></Form.Item></Col>
            </Row>
          </>
        )}
        {modal.type === "weekly" && (
          <>
            <Form.Item name="week_range" label="所属周"><DatePicker.RangePicker style={{ width: "100%" }} /></Form.Item>
            <Row gutter={12}>
              <Col xs={24} md={12}><Form.Item name="stage_goal_id" label="关联阶段目标"><Select allowClear placeholder="不绑定，作为独立周任务" options={stageGoals.map((goal) => ({ value: goal.id, label: goal.title }))} /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="deadline" label="DDL"><DatePicker allowClear style={{ width: "100%" }} /></Form.Item></Col>
            </Row>
            <Row gutter={12}>
              <Col xs={24} md={8}><Form.Item name="status" label="状态"><Select options={STATUS_OPTIONS} /></Form.Item></Col>
              <Col xs={24} md={8}><Form.Item name="priority" label="优先级"><Select options={PRIORITY_OPTIONS} /></Form.Item></Col>
              <Col xs={24} md={8}><Form.Item name="execute_mode" label="执行方式"><Select options={EXECUTE_MODE_OPTIONS} /></Form.Item></Col>
            </Row>
            {executeMode === "scheduled_day" && (
              <Form.Item name="execute_date" label="执行日期" rules={[{ required: true, message: "请选择执行日期" }, { validator: validateDateInWeek }]}>
                <DatePicker style={{ width: "100%" }} />
              </Form.Item>
            )}
            {executeMode === "split_to_days" && (
              <Form.Item name="split_range" label="拆解到这些日期" rules={[{ required: true, message: "请选择拆解日期范围" }, { validator: validateSplitRange }]}>
                <DatePicker.RangePicker style={{ width: "100%" }} />
              </Form.Item>
            )}
          </>
        )}
        {modal.type === "stage" && (
          <>
            <Form.Item name="stage_range" label="阶段周期"><DatePicker.RangePicker style={{ width: "100%" }} /></Form.Item>
            <Form.Item name="target_text" label="成果描述"><Input placeholder="例如：行测稳定 70+，申论每周 3 篇" /></Form.Item>
            <Row gutter={12}>
              <Col xs={24} md={12}><Form.Item name="status" label="状态"><Select options={STAGE_STATUS_OPTIONS} /></Form.Item></Col>
              <Col xs={24} md={12}><Form.Item name="progress_mode" label="进度方式"><Select options={PROGRESS_MODE_OPTIONS} /></Form.Item></Col>
            </Row>
          </>
        )}
        <Form.Item name="note" label="备注"><Input.TextArea rows={3} /></Form.Item>
      </Form>
    </Modal>
  );
}

function TaskMeta({ task }) {
  const meta = [];
  if (getTaskDate(task)) meta.push(`执行 ${formatDate(getTaskDate(task))}`);
  if (getDeadlineDate(task)) meta.push(`DDL ${formatDate(getDeadlineDate(task))}`);
  if (task.estimated_minutes) meta.push(`${task.estimated_minutes} 分钟`);
  if (task.note) meta.push(task.note);
  if (!meta.length) return null;
  return <div className="dashboard-task-meta-line">{meta.join(" · ")}</div>;
}

function CurrentWeeklyList({ tasks }) {
  if (!tasks.length) return <div className="dashboard-stage-current-empty">暂无当前推进项</div>;
  return (
    <div className="dashboard-stage-current">
      {tasks.slice(0, 3).map((task) => (
        <span key={task.id}>{task.title}</span>
      ))}
    </div>
  );
}
